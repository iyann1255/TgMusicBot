/*
 * TgMusicBot - Telegram Music Bot
 *  Copyright (c) 2025-2026 Ashok Shau
 *
 *  Licensed under GNU GPL v3
 *  See https://github.com/AshokShau/TgMusicBot
 */

package vc

import (
	"fmt"
	"strings"
	"time"

	"ashokshau/tgmusic/src/core/cache"

	"github.com/amarnathcjd/gogram/telegram"
)

// LeaveAll makes the bot leave all groups and channels it's currently in.
func (c *TelegramCalls) LeaveAll() (int, error) {
	leftCount := 0

	for _, call := range c.uBContext {
		userBot := call.App

		dialogs, err := userBot.GetDialogs(&telegram.DialogOptions{
			Limit:            -1,
			SleepThresholdMs: 20,
		})
		if err != nil {
			return leftCount, fmt.Errorf("failed to get dialogs: %w", err)
		}

		logger.Info("found dialogs",
			"user", userBot.Me().FirstName,
			"count", len(dialogs),
		)

		activeChats := make(map[int64]bool)
		for _, id := range cache.ChatCache.GetActiveChats() {
			activeChats[id] = true
		}

		for _, d := range dialogs {
			peer := d.Peer
			var chatID int64

			switch p := peer.(type) {
			case *telegram.PeerChannel:
				chatID = p.ChannelID
			case *telegram.PeerChat:
				chatID = p.ChatID
			case *telegram.PeerUser:
				continue
			default:
				continue
			}

			if chatID == 0 || activeChats[chatID] {
				continue
			}

			for {
				err = userBot.LeaveChannel(chatID)
				if err == nil {
					leftCount++
					break
				}

				if strings.Contains(err.Error(), "USER_NOT_PARTICIPANT") ||
					strings.Contains(err.Error(), "CHANNEL_PRIVATE") {
					break
				}

				wait := telegram.GetFloodWait(err)
				if wait > 0 {
					logger.Warn("flood wait",
						"chat_id", chatID,
						"seconds", wait,
					)
					time.Sleep(time.Duration(wait+30) * time.Second)
					continue
				}

				logger.Warn("leave failed",
					"chat_id", chatID,
					"error", err,
				)
				break
			}

			time.Sleep(3 * time.Second)
		}
	}

	return leftCount, nil
}
