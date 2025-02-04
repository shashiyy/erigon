// Copyright 2024 The Erigon Authors
// This file is part of Erigon.
//
// Erigon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Erigon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Erigon. If not, see <http://www.gnu.org/licenses/>.

package sync

import (
	"context"
	"errors"

	"github.com/erigontech/erigon-lib/log/v3"

	"github.com/erigontech/erigon/core/types"
	"github.com/erigontech/erigon/polygon/p2p"
)

type heimdallSynchronizer interface {
	SynchronizeCheckpoints(ctx context.Context) error
	SynchronizeMilestones(ctx context.Context) error
	SynchronizeSpans(ctx context.Context, blockNum uint64) error
}

type bridgeSynchronizer interface {
	Synchronize(ctx context.Context, blockNum uint64) error
}

type Sync struct {
	store             Store
	execution         ExecutionClient
	milestoneVerifier WaypointHeadersVerifier
	blocksVerifier    BlocksVerifier
	p2pService        p2p.Service
	blockDownloader   BlockDownloader
	ccBuilderFactory  CanonicalChainBuilderFactory
	heimdallSync      heimdallSynchronizer
	bridgeSync        bridgeSynchronizer
	events            <-chan Event
	logger            log.Logger
}

func NewSync(
	store Store,
	execution ExecutionClient,
	milestoneVerifier WaypointHeadersVerifier,
	blocksVerifier BlocksVerifier,
	p2pService p2p.Service,
	blockDownloader BlockDownloader,
	ccBuilderFactory CanonicalChainBuilderFactory,
	heimdallSync heimdallSynchronizer,
	bridgeSync bridgeSynchronizer,
	events <-chan Event,
	logger log.Logger,
) *Sync {
	return &Sync{
		store:             store,
		execution:         execution,
		milestoneVerifier: milestoneVerifier,
		blocksVerifier:    blocksVerifier,
		p2pService:        p2pService,
		blockDownloader:   blockDownloader,
		ccBuilderFactory:  ccBuilderFactory,
		heimdallSync:      heimdallSync,
		bridgeSync:        bridgeSync,
		events:            events,
		logger:            logger,
	}
}

func (s *Sync) commitExecution(ctx context.Context, newTip *types.Header, finalizedHeader *types.Header) error {
	if err := s.store.Flush(ctx); err != nil {
		return err
	}

	blockNum := newTip.Number.Uint64()

	if err := s.heimdallSync.SynchronizeSpans(ctx, blockNum); err != nil {
		return err
	}

	if err := s.bridgeSync.Synchronize(ctx, blockNum); err != nil {
		return err
	}

	return s.execution.UpdateForkChoice(ctx, newTip, finalizedHeader)
}

func (s *Sync) handleMilestoneMismatch(ctx context.Context, ccBuilder CanonicalChainBuilder) error {
	// the milestone doesn't correspond to the tip of the chain
	// unwind to the previous verified milestone
	oldTip := ccBuilder.Root()
	oldTipNum := oldTip.Number.Uint64()

	if err := s.execution.UpdateForkChoice(ctx, oldTip, oldTip); err != nil {
		return err
	}

	newTip, err := s.blockDownloader.DownloadBlocksUsingMilestones(ctx, oldTipNum)
	if err != nil {
		return err
	}
	if newTip == nil {
		return errors.New("sync.Sync.handleMilestoneMismatch: unexpected to have no milestone headers since the last milestone after receiving a new milestone event")
	}

	if err = s.commitExecution(ctx, newTip, newTip); err != nil {
		return err
	}

	ccBuilder.Reset(newTip)

	return nil
}

func (s *Sync) onMilestoneEvent(
	ctx context.Context,
	event EventNewMilestone,
	ccBuilder CanonicalChainBuilder,
) error {
	milestone := event
	if milestone.EndBlock().Uint64() <= ccBuilder.Root().Number.Uint64() {
		return nil
	}

	milestoneHeaders := ccBuilder.HeadersInRange(milestone.StartBlock().Uint64(), milestone.Length())
	err := s.milestoneVerifier(milestone, milestoneHeaders)
	if err != nil {
		s.logger.Debug(
			syncLogPrefix("onMilestoneEvent: local chain tip does not match the milestone, unwinding to the previous verified milestone"),
			"err", err,
		)

		return s.handleMilestoneMismatch(ctx, ccBuilder)
	}

	return ccBuilder.Prune(milestone.EndBlock().Uint64())
}

func (s *Sync) onNewBlockEvent(
	ctx context.Context,
	event EventNewBlock,
	ccBuilder CanonicalChainBuilder,
) error {
	newBlockHeader := event.NewBlock.Header()
	newBlockHeaderNum := newBlockHeader.Number.Uint64()
	rootNum := ccBuilder.Root().Number.Uint64()
	if newBlockHeaderNum <= rootNum {
		return nil
	}

	var newBlocks []*types.Block
	var err error
	if ccBuilder.ContainsHash(newBlockHeader.ParentHash) {
		newBlocks = []*types.Block{event.NewBlock}
	} else {
		blocks, err := s.p2pService.FetchBlocks(ctx, rootNum, newBlockHeaderNum+1, event.PeerId)
		if err != nil {
			if (p2p.ErrIncompleteHeaders{}).Is(err) || (p2p.ErrMissingBodies{}).Is(err) {
				s.logger.Debug(
					syncLogPrefix("onNewBlockEvent: failed to fetch complete blocks, ignoring event"),
					"err", err,
					"peerId", event.PeerId,
					"lastBlockNum", newBlockHeaderNum,
				)

				return nil
			}

			return err
		}

		newBlocks = blocks.Data
	}

	if err := s.blocksVerifier(newBlocks); err != nil {
		s.logger.Debug(syncLogPrefix("onNewBlockEvent: invalid new block event from peer, penalizing and ignoring"), "err", err)

		if err = s.p2pService.Penalize(ctx, event.PeerId); err != nil {
			s.logger.Debug(syncLogPrefix("onNewBlockEvent: issue with penalizing peer"), "err", err)
		}

		return nil
	}

	newHeaders := make([]*types.Header, len(newBlocks))
	for i, block := range newBlocks {
		newHeaders[i] = block.HeaderNoCopy()
	}

	oldTip := ccBuilder.Tip()
	if err = ccBuilder.Connect(ctx, newHeaders); err != nil {
		s.logger.Debug(syncLogPrefix("onNewBlockEvent: couldn't connect a header to the local chain tip, ignoring"), "err", err)
		return nil
	}

	newTip := ccBuilder.Tip()
	if newTip != oldTip {
		if err = s.execution.InsertBlocks(ctx, newBlocks); err != nil {
			return err
		}

		if err = s.execution.UpdateForkChoice(ctx, newTip, ccBuilder.Root()); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sync) onNewBlockHashesEvent(
	ctx context.Context,
	event EventNewBlockHashes,
	ccBuilder CanonicalChainBuilder,
) error {
	for _, headerHashNum := range event.NewBlockHashes {
		if (headerHashNum.Number <= ccBuilder.Root().Number.Uint64()) || ccBuilder.ContainsHash(headerHashNum.Hash) {
			continue
		}

		newBlocks, err := s.p2pService.FetchBlocks(ctx, headerHashNum.Number, headerHashNum.Number+1, event.PeerId)
		if err != nil {
			if (p2p.ErrIncompleteHeaders{}).Is(err) || (p2p.ErrMissingBodies{}).Is(err) {
				s.logger.Debug(
					syncLogPrefix("onNewBlockHashesEvent: failed to fetch complete blocks, ignoring event"),
					"err", err,
					"peerId", event.PeerId,
					"lastBlockNum", headerHashNum.Number,
				)

				continue
			}

			return err
		}

		newBlockEvent := EventNewBlock{
			NewBlock: newBlocks.Data[0],
			PeerId:   event.PeerId,
		}

		err = s.onNewBlockEvent(ctx, newBlockEvent, ccBuilder)
		if err != nil {
			return err
		}
	}
	return nil
}

//
// TODO (subsequent PRs) - unit test initial sync + on new event cases
//

func (s *Sync) Run(ctx context.Context) error {
	s.logger.Debug(syncLogPrefix("running sync component"))

	tip, err := s.syncToTip(ctx)
	if err != nil {
		return err
	}

	ccBuilder := s.ccBuilderFactory(tip)

	for {
		select {
		case event := <-s.events:
			switch event.Type {
			case EventTypeNewMilestone:
				if err = s.onMilestoneEvent(ctx, event.AsNewMilestone(), ccBuilder); err != nil {
					return err
				}
			case EventTypeNewBlock:
				if err = s.onNewBlockEvent(ctx, event.AsNewBlock(), ccBuilder); err != nil {
					return err
				}
			case EventTypeNewBlockHashes:
				if err = s.onNewBlockHashesEvent(ctx, event.AsNewBlockHashes(), ccBuilder); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Sync) syncToTip(ctx context.Context) (*types.Header, error) {
	tip, err := s.execution.CurrentHeader(ctx)
	if err != nil {
		return nil, err
	}

	tip, err = s.syncToTipUsingCheckpoints(ctx, tip)
	if err != nil {
		return nil, err
	}

	tip, err = s.syncToTipUsingMilestones(ctx, tip)
	if err != nil {
		return nil, err
	}

	return tip, nil
}

func (s *Sync) syncToTipUsingCheckpoints(ctx context.Context, tip *types.Header) (*types.Header, error) {
	return s.sync(ctx, tip, func(ctx context.Context, startBlockNum uint64) (*types.Header, error) {
		err := s.heimdallSync.SynchronizeCheckpoints(ctx)
		if err != nil {
			return nil, err
		}

		return s.blockDownloader.DownloadBlocksUsingCheckpoints(ctx, startBlockNum)
	})
}

func (s *Sync) syncToTipUsingMilestones(ctx context.Context, tip *types.Header) (*types.Header, error) {
	return s.sync(ctx, tip, func(ctx context.Context, startBlockNum uint64) (*types.Header, error) {
		err := s.heimdallSync.SynchronizeMilestones(ctx)
		if err != nil {
			return nil, err
		}

		return s.blockDownloader.DownloadBlocksUsingMilestones(ctx, startBlockNum)
	})
}

func (s *Sync) sync(
	ctx context.Context,
	tip *types.Header,
	tipDownloader func(ctx context.Context, startBlockNum uint64) (*types.Header, error),
) (*types.Header, error) {
	for {
		newTip, err := tipDownloader(ctx, tip.Number.Uint64()+1)
		if err != nil {
			return nil, err
		}

		if newTip == nil {
			// we've reached the tip
			break
		}

		tip = newTip
		if err = s.commitExecution(ctx, tip, tip); err != nil {
			return nil, err
		}
	}

	return tip, nil
}
