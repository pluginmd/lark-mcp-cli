// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/larksuite/cli/shortcuts/common"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

// ImCardSend renders a high-level YAML CardSpec into Feishu's
// interactive-card JSON and sends it. Wraps both `messages.create`
// (default) and `messages.reply` (when --reply-to is set) so callers
// don't have to choose a different shortcut depending on whether
// they're starting a new card or replying with one.
var ImCardSend = common.Shortcut{
	Service:     "im",
	Command:     "+card-send",
	Description: "Render a YAML card spec into a Feishu interactive card and send/reply; user/bot; supports collapsible panels, action buttons, list items, columns, summary preview; --print-json compiles without sending",
	Risk:        "write",
	// Scopes are explicitly empty at the unconditional pre-flight
	// level because --print-json is a pure offline compile path that
	// hits no API. The actual scope requirements are declared as
	// Conditional so they still surface in metadata, auth hints, and
	// scope-diagnosis output — the real send path fails fast with
	// the same scope error if permission is missing.
	// The empty []string{} (not nil) satisfies the repo-wide
	// TestAllShortcutsScopesNotNil invariant in shortcuts/register_test.go.
	Scopes:                []string{},
	ConditionalScopes:     []string{"im:message:send_as_bot"},
	ConditionalUserScopes: []string{"im:message.send_as_user", "im:message"},
	ConditionalBotScopes:  []string{"im:message:send_as_bot"},
	AuthTypes:             []string{"bot", "user"},
	Flags: []common.Flag{
		{Name: "chat-id", Desc: "target chat ID (oc_xxx); mutually exclusive with --user-id and --reply-to"},
		{Name: "user-id", Desc: "target user open_id (ou_xxx); mutually exclusive with --chat-id and --reply-to"},
		{Name: "reply-to", Desc: "reply to this message ID (om_xxx); mutually exclusive with --chat-id and --user-id"},
		{Name: "in-thread", Type: "bool", Desc: "with --reply-to: reply in thread (message appears in thread stream)"},
		{Name: "spec", Required: true, Desc: "YAML card spec; use @path to read from file, - to read from stdin, or pass inline YAML", Input: []string{common.File, common.Stdin}},
		{Name: "idempotency-key", Desc: "idempotency key (prevents duplicate sends)"},
		{Name: "print-json", Type: "bool", Desc: "compile the spec and print the Feishu card JSON to stdout without sending"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		chatID := runtime.Str("chat-id")
		userID := runtime.Str("user-id")
		replyTo := runtime.Str("reply-to")
		inThread := runtime.Bool("in-thread")
		spec := runtime.Str("spec")
		idem := runtime.Str("idempotency-key")

		// Best-effort compile so the dry-run preview shows the real card body.
		content := "<spec compile error — see Validate output>"
		if compiled, err := Compile([]byte(spec)); err == nil {
			content = compiled
		}

		body := map[string]interface{}{
			"msg_type": "interactive",
			"content":  content,
		}
		if idem != "" {
			body["uuid"] = idem
		}

		d := common.NewDryRunAPI().Desc("send interactive card")
		if replyTo != "" {
			if inThread {
				body["reply_in_thread"] = true
			}
			return d.
				POST("/open-apis/im/v1/messages/:message_id/reply").
				Body(body).
				Set("message_id", replyTo)
		}
		receiveIDType := "chat_id"
		receiveID := chatID
		if userID != "" {
			receiveIDType = "open_id"
			receiveID = userID
		}
		body["receive_id"] = receiveID
		return d.
			POST("/open-apis/im/v1/messages").
			Params(map[string]interface{}{"receive_id_type": receiveIDType}).
			Body(body)
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		chatID := runtime.Str("chat-id")
		userID := runtime.Str("user-id")
		replyTo := runtime.Str("reply-to")
		inThread := runtime.Bool("in-thread")
		spec := runtime.Str("spec")
		printJSON := runtime.Bool("print-json")

		// Exactly one destination — unless --print-json is set, in which case
		// no destination is needed because we never call the API.
		if !printJSON {
			if err := common.ExactlyOne(runtime, "chat-id", "user-id", "reply-to"); err != nil {
				return err
			}
		}
		if chatID != "" {
			if _, err := common.ValidateChatID(chatID); err != nil {
				return err
			}
		}
		if userID != "" {
			if _, err := common.ValidateUserID(userID); err != nil {
				return err
			}
		}
		if replyTo != "" && !strings.HasPrefix(replyTo, "om_") {
			return common.FlagErrorf("--reply-to: must start with om_ (got %q)", replyTo)
		}
		if inThread && replyTo == "" {
			return common.FlagErrorf("--in-thread requires --reply-to")
		}
		if spec == "" {
			return common.FlagErrorf("--spec is required (inline YAML, @path, or -)")
		}
		// Compile up front so YAML errors surface before any API call.
		if _, err := Compile([]byte(spec)); err != nil {
			return common.FlagErrorf("--spec: %v", err)
		}
		return nil
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		chatID := runtime.Str("chat-id")
		userID := runtime.Str("user-id")
		replyTo := runtime.Str("reply-to")
		inThread := runtime.Bool("in-thread")
		spec := runtime.Str("spec")
		idem := runtime.Str("idempotency-key")
		printJSON := runtime.Bool("print-json")

		content, err := Compile([]byte(spec))
		if err != nil {
			return common.FlagErrorf("--spec: %v", err)
		}

		if printJSON {
			// Pretty-print so it can be eyeballed/diffed. Output is the only
			// stdout we emit; metadata goes through runtime.Out machinery.
			var pretty map[string]any
			_ = json.Unmarshal([]byte(content), &pretty)
			runtime.OutRaw(pretty, nil)
			return nil
		}

		body := map[string]interface{}{
			"msg_type": "interactive",
			"content":  content,
		}
		if idem != "" {
			body["uuid"] = idem
		}

		var apiPath string
		var query larkcore.QueryParams
		if replyTo != "" {
			if inThread {
				body["reply_in_thread"] = true
			}
			apiPath = "/open-apis/im/v1/messages/" + replyTo + "/reply"
		} else {
			receiveIDType := "chat_id"
			receiveID := chatID
			if userID != "" {
				receiveIDType = "open_id"
				receiveID = userID
			}
			body["receive_id"] = receiveID
			apiPath = "/open-apis/im/v1/messages"
			query = larkcore.QueryParams{"receive_id_type": []string{receiveIDType}}
		}

		resData, err := runtime.DoAPIJSON(http.MethodPost, apiPath, query, body)
		if err != nil {
			return err
		}
		runtime.Out(map[string]interface{}{
			"message_id":  resData["message_id"],
			"chat_id":     resData["chat_id"],
			"create_time": common.FormatTimeWithSeconds(resData["create_time"]),
		}, nil)
		return nil
	},
}
