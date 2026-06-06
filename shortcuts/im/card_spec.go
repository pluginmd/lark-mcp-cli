// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// CardSpec is the high-level YAML schema users write. The compiler walks
// it and produces Feishu's interactive-card JSON (schema 2.0). Why a
// typed tree rather than yaml.MapSlice: each kind gets a precise error
// pointing at the offending field, which is the whole reason for having
// a DSL instead of asking users to hand-roll the 200-line raw JSON.
type CardSpec struct {
	Header      *CardHeader       `yaml:"header,omitempty"`
	Summary     string            `yaml:"summary,omitempty"`
	SummaryI18n map[string]string `yaml:"summary_i18n,omitempty"`
	Locales     []string          `yaml:"locales,omitempty"`
	Wide        *bool             `yaml:"wide,omitempty"`
	UpdateMulti *bool             `yaml:"update_multi,omitempty"`
	Elements    []CardElement     `yaml:"elements"`
}

type CardHeader struct {
	Title         string            `yaml:"title"`
	TitleI18n     map[string]string `yaml:"title_i18n,omitempty"`
	TitleColor    string            `yaml:"title_color,omitempty"`
	Subtitle      string            `yaml:"subtitle,omitempty"`
	SubtitleI18n  map[string]string `yaml:"subtitle_i18n,omitempty"`
	SubtitleColor string            `yaml:"subtitle_color,omitempty"`
	Template      string            `yaml:"template,omitempty"`   // colour token
	Icon          string            `yaml:"icon,omitempty"`       // "standard:<token>" | "custom:<key>"
	IconColor     string            `yaml:"icon_color,omitempty"` // colour for standard_icon
	IconSize      string            `yaml:"icon_size,omitempty"`  // e.g. "16px 16px"
}

// CardElement uses a UnmarshalYAML hook to dispatch on the discriminator
// key (md / hr / actions / panel / …). The "kind" + "data" split keeps
// each compile branch independent.
type CardElement struct {
	Kind string
	Data any
}

type cardMarkdown struct {
	Content string            `yaml:"content"`
	Size    string            `yaml:"size,omitempty"`     // notation | normal_v2
	Align   string            `yaml:"align,omitempty"`    // left | center | right
	Margin  string            `yaml:"margin,omitempty"`   // "8px 0px 0px 0px"
	I18n    map[string]string `yaml:"i18n,omitempty"`     // per-locale alternative content
	NoStyle bool              `yaml:"no_style,omitempty"` // opt out of optimizeMarkdownStyle
}

// cardDiv is openclaw-lark's primary primitive for "icon + text" status
// rows that markdown can't compose (markdown has no sized inline icon).
// Mirrors the {tag:'div', icon:{...}, text:{...}} shape used throughout
// builder.ts.
type cardDiv struct {
	Text      string            `yaml:"text"`
	TextI18n  map[string]string `yaml:"text_i18n,omitempty"`
	TextSize  string            `yaml:"text_size,omitempty"` // notation | normal_v2
	TextColor string            `yaml:"text_color,omitempty"`
	TextTag   string            `yaml:"text_tag,omitempty"` // plain_text (default) | lark_md
	Icon      string            `yaml:"icon,omitempty"`     // "standard:<token>" | "custom:<key>"
	IconColor string            `yaml:"icon_color,omitempty"`
	IconSize  string            `yaml:"icon_size,omitempty"`
	Margin    string            `yaml:"margin,omitempty"`
}

type cardActions struct {
	Items  []cardButton `yaml:"items"`
	Layout string       `yaml:"layout,omitempty"` // equal | inline
}

type cardButton struct {
	Text  string `yaml:"text"`
	Value string `yaml:"value,omitempty"`
	Style string `yaml:"style,omitempty"` // primary | danger | default
	URL   string `yaml:"url,omitempty"`   // when set, button opens URL instead of callback
}

type cardItem struct {
	Text string     `yaml:"text"`
	Btn  cardButton `yaml:"btn"`
}

type cardSelect struct {
	Placeholder string             `yaml:"placeholder,omitempty"`
	Value       string             `yaml:"value,omitempty"`
	Initial     string             `yaml:"initial,omitempty"`
	Options     []cardSelectOption `yaml:"options"`
}

type cardSelectOption struct {
	Text  string `yaml:"text"`
	Value string `yaml:"value"`
}

type cardPanel struct {
	Title      string            `yaml:"title"`
	TitleI18n  map[string]string `yaml:"title_i18n,omitempty"`
	TitleColor string            `yaml:"title_color,omitempty"`
	TitleTag   string            `yaml:"title_tag,omitempty"` // plain_text (default) | markdown
	Border     string            `yaml:"border,omitempty"`    // grey | blue | red | green | orange
	Expanded   bool              `yaml:"expanded,omitempty"`
	Elements   []CardElement     `yaml:"elements"`
}

type cardColumns struct {
	Flex  string       `yaml:"flex,omitempty"` // bisect | flow
	Items []cardColumn `yaml:"items"`
}

type cardColumn struct {
	Weight   any           `yaml:"weight,omitempty"` // int or "auto"
	Elements []CardElement `yaml:"elements"`
}

type cardImage struct {
	Key     string `yaml:"key"`
	Alt     string `yaml:"alt,omitempty"`
	Compact bool   `yaml:"compact,omitempty"`
}

// validTemplates / validBorders are validated up front so a typo doesn't
// silently render as default colour at Feishu's end.
var validTemplates = map[string]struct{}{
	"blue": {}, "turquoise": {}, "red": {}, "green": {}, "orange": {},
	"wathet": {}, "indigo": {}, "carmine": {}, "grey": {},
}

var validBorders = map[string]struct{}{
	"grey": {}, "blue": {}, "red": {}, "green": {}, "orange": {},
}

var validButtonStyles = map[string]struct{}{
	"primary": {}, "danger": {}, "default": {},
}

// validMdSizes mirrors what both reference bridges actually emit
// (Lark fork's run-renderer.ts, cc-connect's CardMarkdown). "heading"
// and "title" are documented on Feishu but never used by these bridges
// so we don't whitelist them here — users who need them can fall
// through to the `raw` element kind.
var validMdSizes = map[string]struct{}{
	"notation": {}, "normal_v2": {},
}

// validTextColors mirrors the colour tokens openclaw-lark uses on
// plain_text elements (header titles, panel titles, footer text).
// Feishu accepts more named colours in CSS-style content, but for
// element-level text_color this is the verified safe list.
var validTextColors = map[string]struct{}{
	"grey": {}, "red": {}, "green": {}, "blue": {},
	"turquoise": {}, "orange": {}, "purple": {}, "yellow": {},
}

var validTextAligns = map[string]struct{}{
	"left": {}, "center": {}, "right": {},
}

var validTextTagsForDiv = map[string]struct{}{
	"plain_text": {}, "lark_md": {},
}

var validTextTagsForPanelTitle = map[string]struct{}{
	"plain_text": {}, "markdown": {},
}

// UnmarshalYAML dispatches each list item by its discriminator key.
// One sealed switch keeps the spec closed — unknown kinds error loudly.
func (e *CardElement) UnmarshalYAML(node *yaml.Node) error {
	// Scalar shorthand: `- hr`
	if node.Kind == yaml.ScalarNode && node.Value == "hr" {
		e.Kind = "hr"
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: element must be a mapping or the scalar 'hr', got %s", node.Line, kindName(node.Kind))
	}
	if len(node.Content) < 2 {
		return fmt.Errorf("line %d: empty element", node.Line)
	}
	// First key is the discriminator. Single-key mapping is the canonical form;
	// extra keys are allowed for actions { items, layout } and similar combos
	// that already carry their own typed struct.
	disc := node.Content[0].Value

	switch disc {
	case "md":
		valNode := node.Content[1]
		md := cardMarkdown{}
		if valNode.Kind == yaml.ScalarNode {
			md.Content = valNode.Value
		} else if err := valNode.Decode(&md); err != nil {
			return fmt.Errorf("line %d: invalid md: %w", node.Line, err)
		}
		if md.Content == "" {
			return fmt.Errorf("line %d: md.content is empty", node.Line)
		}
		if md.Size != "" {
			if _, ok := validMdSizes[md.Size]; !ok {
				return fmt.Errorf("line %d: invalid md.size %q (allowed: notation, normal_v2)", node.Line, md.Size)
			}
		}
		if md.Align != "" {
			if _, ok := validTextAligns[md.Align]; !ok {
				return fmt.Errorf("line %d: invalid md.align %q (allowed: left, center, right)", node.Line, md.Align)
			}
		}
		e.Kind, e.Data = "md", md

	case "hr":
		e.Kind = "hr"

	case "note":
		var s string
		if err := node.Content[1].Decode(&s); err != nil {
			return fmt.Errorf("line %d: note must be a string: %w", node.Line, err)
		}
		if s == "" {
			return fmt.Errorf("line %d: note is empty", node.Line)
		}
		e.Kind, e.Data = "note", s

	case "actions":
		a := cardActions{}
		// items can be inline under "actions:" OR nested under "actions: { items: [...], layout: ... }".
		if node.Content[1].Kind == yaml.SequenceNode {
			if err := node.Content[1].Decode(&a.Items); err != nil {
				return fmt.Errorf("line %d: invalid actions: %w", node.Line, err)
			}
			// Look for an optional sibling "layout" key on the parent.
			for i := 2; i+1 < len(node.Content); i += 2 {
				if node.Content[i].Value == "layout" {
					a.Layout = node.Content[i+1].Value
				}
			}
		} else if err := node.Content[1].Decode(&a); err != nil {
			return fmt.Errorf("line %d: invalid actions: %w", node.Line, err)
		}
		if err := validateButtons(a.Items, node.Line); err != nil {
			return err
		}
		if a.Layout != "" && a.Layout != "equal" && a.Layout != "inline" {
			return fmt.Errorf("line %d: invalid actions.layout %q (allowed: equal, inline)", node.Line, a.Layout)
		}
		e.Kind, e.Data = "actions", a

	case "item":
		var it cardItem
		if err := node.Content[1].Decode(&it); err != nil {
			return fmt.Errorf("line %d: invalid item: %w", node.Line, err)
		}
		if it.Text == "" {
			return fmt.Errorf("line %d: item.text is empty", node.Line)
		}
		if it.Btn.Text == "" {
			return fmt.Errorf("line %d: item.btn.text is empty", node.Line)
		}
		if err := validateButtons([]cardButton{it.Btn}, node.Line); err != nil {
			return err
		}
		e.Kind, e.Data = "item", it

	case "select":
		var s cardSelect
		if err := node.Content[1].Decode(&s); err != nil {
			return fmt.Errorf("line %d: invalid select: %w", node.Line, err)
		}
		if len(s.Options) == 0 {
			return fmt.Errorf("line %d: select.options is empty", node.Line)
		}
		e.Kind, e.Data = "select", s

	case "panel":
		var p cardPanel
		if err := node.Content[1].Decode(&p); err != nil {
			return fmt.Errorf("line %d: invalid panel: %w", node.Line, err)
		}
		if p.Title == "" {
			return fmt.Errorf("line %d: panel.title is empty", node.Line)
		}
		if p.Border != "" {
			if _, ok := validBorders[p.Border]; !ok {
				return fmt.Errorf("line %d: invalid panel.border %q (allowed: grey, blue, red, green, orange)", node.Line, p.Border)
			}
		}
		if p.TitleColor != "" {
			if _, ok := validTextColors[p.TitleColor]; !ok {
				return fmt.Errorf("line %d: invalid panel.title_color %q (allowed: grey, red, green, blue, turquoise, orange, purple, yellow)", node.Line, p.TitleColor)
			}
		}
		if p.TitleTag != "" {
			if _, ok := validTextTagsForPanelTitle[p.TitleTag]; !ok {
				return fmt.Errorf("line %d: invalid panel.title_tag %q (allowed: plain_text, markdown)", node.Line, p.TitleTag)
			}
		}
		e.Kind, e.Data = "panel", p

	case "div":
		var dv cardDiv
		if err := node.Content[1].Decode(&dv); err != nil {
			return fmt.Errorf("line %d: invalid div: %w", node.Line, err)
		}
		if dv.Text == "" {
			return fmt.Errorf("line %d: div.text is empty", node.Line)
		}
		if dv.TextSize != "" {
			if _, ok := validMdSizes[dv.TextSize]; !ok {
				return fmt.Errorf("line %d: invalid div.text_size %q (allowed: notation, normal_v2)", node.Line, dv.TextSize)
			}
		}
		if dv.TextColor != "" {
			if _, ok := validTextColors[dv.TextColor]; !ok {
				return fmt.Errorf("line %d: invalid div.text_color %q (allowed: grey, red, green, blue, turquoise, orange, purple, yellow)", node.Line, dv.TextColor)
			}
		}
		if dv.TextTag != "" {
			if _, ok := validTextTagsForDiv[dv.TextTag]; !ok {
				return fmt.Errorf("line %d: invalid div.text_tag %q (allowed: plain_text, lark_md)", node.Line, dv.TextTag)
			}
		}
		if dv.IconColor != "" {
			if _, ok := validTextColors[dv.IconColor]; !ok {
				return fmt.Errorf("line %d: invalid div.icon_color %q (allowed: grey, red, green, blue, turquoise, orange, purple, yellow)", node.Line, dv.IconColor)
			}
		}
		if dv.Icon != "" {
			if err := validateIconSpec("div.icon", dv.Icon); err != nil {
				return fmt.Errorf("line %d: %v", node.Line, err)
			}
		}
		e.Kind, e.Data = "div", dv

	case "columns":
		var c cardColumns
		if err := node.Content[1].Decode(&c); err != nil {
			return fmt.Errorf("line %d: invalid columns: %w", node.Line, err)
		}
		if len(c.Items) == 0 {
			return fmt.Errorf("line %d: columns.items is empty", node.Line)
		}
		if c.Flex != "" && c.Flex != "bisect" && c.Flex != "flow" && c.Flex != "stretch" {
			return fmt.Errorf("line %d: invalid columns.flex %q (allowed: bisect, flow, stretch)", node.Line, c.Flex)
		}
		e.Kind, e.Data = "columns", c

	case "image":
		var img cardImage
		if err := node.Content[1].Decode(&img); err != nil {
			return fmt.Errorf("line %d: invalid image: %w", node.Line, err)
		}
		if img.Key == "" {
			return fmt.Errorf("line %d: image.key is empty", node.Line)
		}
		if !strings.HasPrefix(img.Key, "img_") {
			return fmt.Errorf("line %d: image.key %q must start with 'img_' — upload first via `lark-cli im images +upload` to obtain a valid key", node.Line, img.Key)
		}
		e.Kind, e.Data = "image", img

	case "raw":
		// Direct passthrough — the user knows what they're doing.
		var raw map[string]any
		if err := node.Content[1].Decode(&raw); err != nil {
			return fmt.Errorf("line %d: invalid raw element: %w", node.Line, err)
		}
		if _, ok := raw["tag"]; !ok {
			return fmt.Errorf("line %d: raw element must include a 'tag' field", node.Line)
		}
		e.Kind, e.Data = "raw", raw

	default:
		return fmt.Errorf("line %d: unknown element kind %q (allowed: md, hr, note, actions, item, select, panel, columns, image, div, raw)", node.Line, disc)
	}
	return nil
}

func kindName(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	}
	return "unknown"
}

func validateButtons(btns []cardButton, line int) error {
	if len(btns) == 0 {
		return fmt.Errorf("line %d: actions/item must declare at least one button", line)
	}
	for i, b := range btns {
		if b.Text == "" {
			return fmt.Errorf("line %d: button[%d].text is empty", line, i)
		}
		if b.Style != "" {
			if _, ok := validButtonStyles[b.Style]; !ok {
				return fmt.Errorf("line %d: button[%d].style %q invalid (allowed: primary, danger, default)", line, i, b.Style)
			}
		}
		if b.URL == "" && b.Value == "" {
			return fmt.Errorf("line %d: button[%d] must set either value (callback) or url (link)", line, i)
		}
	}
	return nil
}

// Compile turns a YAML spec into the Feishu interactive-card JSON string
// that `im.v1.message.create` accepts as `content` when msg_type=interactive.
// The output is always schema 2.0; we don't support v1 because v2.0 covers
// every element v1 had plus collapsible_panel, which is the whole point.
func Compile(specYAML []byte) (string, error) {
	var spec CardSpec
	// Strict decoding so typos at the top level (`templat:` instead of
	// `template:`) raise a precise error at Validate time instead of
	// silently dropping the field — the most common YAML footgun in
	// hand-written specs. Note: this only catches typos in fields the
	// top-level CardSpec / CardHeader / cardMarkdown / etc. structs
	// know about. The element discriminator dispatch already rejects
	// unknown element kinds in UnmarshalYAML.
	dec := yaml.NewDecoder(bytes.NewReader(specYAML))
	dec.KnownFields(true)
	if err := dec.Decode(&spec); err != nil {
		return "", fmt.Errorf("parse spec: %w", err)
	}
	if len(spec.Elements) == 0 {
		return "", fmt.Errorf("spec.elements is empty — a card needs at least one element")
	}
	if spec.Header != nil {
		if spec.Header.Title == "" {
			return "", fmt.Errorf("header.title is required when header is set")
		}
		if spec.Header.Template != "" {
			if _, ok := validTemplates[spec.Header.Template]; !ok {
				return "", fmt.Errorf("invalid header.template %q (allowed: blue, turquoise, red, green, orange, wathet, indigo, carmine, grey)", spec.Header.Template)
			}
		}
		if spec.Header.TitleColor != "" {
			if _, ok := validTextColors[spec.Header.TitleColor]; !ok {
				return "", fmt.Errorf("invalid header.title_color %q (allowed: grey, red, green, blue, turquoise, orange, purple, yellow)", spec.Header.TitleColor)
			}
		}
		if spec.Header.SubtitleColor != "" {
			if _, ok := validTextColors[spec.Header.SubtitleColor]; !ok {
				return "", fmt.Errorf("invalid header.subtitle_color %q (allowed: grey, red, green, blue, turquoise, orange, purple, yellow)", spec.Header.SubtitleColor)
			}
		}
		if spec.Header.IconColor != "" {
			if _, ok := validTextColors[spec.Header.IconColor]; !ok {
				return "", fmt.Errorf("invalid header.icon_color %q (allowed: grey, red, green, blue, turquoise, orange, purple, yellow)", spec.Header.IconColor)
			}
		}
		if spec.Header.Icon != "" {
			if err := validateIconSpec("header.icon", spec.Header.Icon); err != nil {
				return "", err
			}
		}
	}

	// Config starts empty. wide_screen_mode and update_multi are v1
	// fields; in schema 2.0 they're silently accepted but non-canonical
	// (the Lark fork's toCardKit2 translator strips them on v1→v2
	// upgrade). We only emit them when the spec explicitly sets them,
	// keeping the default output clean and idiomatic for v2.0.
	config := map[string]any{}
	if spec.Wide != nil {
		config["wide_screen_mode"] = *spec.Wide
	}
	if spec.UpdateMulti != nil {
		config["update_multi"] = *spec.UpdateMulti
	}
	if spec.Summary != "" {
		summary := map[string]any{"content": spec.Summary}
		if len(spec.SummaryI18n) > 0 {
			summary["i18n_content"] = stringMapToAny(spec.SummaryI18n)
		}
		config["summary"] = summary
	}
	// locales must be declared at config level when i18n_content appears
	// on any element (openclaw-lark sets ['zh_cn', 'en_us'] for bilingual
	// cards). Explicit `locales:` in the spec wins; otherwise we auto-
	// derive it from the union of every i18n map's keys so users don't
	// have to maintain the list manually.
	locales := spec.Locales
	if len(locales) == 0 {
		locales = collectLocales(&spec)
	}
	if len(locales) > 0 {
		config["locales"] = stringSliceToAny(locales)
	}

	card := map[string]any{
		"schema": "2.0",
		"config": config,
	}
	if spec.Header != nil {
		card["header"] = compileHeader(spec.Header)
	}

	elements, err := compileElements(spec.Elements)
	if err != nil {
		return "", err
	}
	card["body"] = map[string]any{"elements": elements}

	out, err := json.Marshal(card)
	if err != nil {
		return "", fmt.Errorf("marshal card: %w", err)
	}
	return string(out), nil
}

func compileHeader(h *CardHeader) map[string]any {
	title := map[string]any{"tag": "plain_text", "content": h.Title}
	if len(h.TitleI18n) > 0 {
		title["i18n_content"] = stringMapToAny(h.TitleI18n)
	}
	if h.TitleColor != "" {
		title["text_color"] = h.TitleColor
	}
	hm := map[string]any{"title": title}

	if h.Subtitle != "" {
		sub := map[string]any{"tag": "plain_text", "content": h.Subtitle}
		if len(h.SubtitleI18n) > 0 {
			sub["i18n_content"] = stringMapToAny(h.SubtitleI18n)
		}
		if h.SubtitleColor != "" {
			sub["text_color"] = h.SubtitleColor
		}
		hm["subtitle"] = sub
	}
	if h.Template != "" {
		hm["template"] = h.Template
	}
	if h.Icon != "" {
		hm["icon"] = compileIconSpec(h.Icon, h.IconColor, h.IconSize)
	}
	return hm
}

// validateIconSpec checks an icon string follows one of the two
// supported prefix forms. Called from Compile() before any compileIconSpec
// site so an unrecognised icon errors loudly at Validate time instead
// of silently dropping (compileIconSpec returns nil for unknown prefix,
// which would render a header/div with no icon — a hard-to-debug
// "where did my icon go" UX failure).
func validateIconSpec(field, spec string) error {
	switch {
	case strings.HasPrefix(spec, "standard:"):
		if strings.TrimPrefix(spec, "standard:") == "" {
			return fmt.Errorf("%s: 'standard:' prefix requires a token (e.g. standard:notice-bell)", field)
		}
	case strings.HasPrefix(spec, "custom:"):
		key := strings.TrimPrefix(spec, "custom:")
		if key == "" {
			return fmt.Errorf("%s: 'custom:' prefix requires an image key (e.g. custom:img_xxx)", field)
		}
		if !strings.HasPrefix(key, "img_") {
			return fmt.Errorf("%s: custom icon key %q must start with 'img_' — upload first via `lark-cli im images +upload`", field, key)
		}
	default:
		return fmt.Errorf("%s: %q has no recognised prefix (use 'standard:<token>' or 'custom:img_xxx')", field, spec)
	}
	return nil
}

// compileIconSpec parses "standard:<token>" or "custom:<img_key>" into
// the Feishu icon element JSON, attaching optional colour and size.
// Returns nil for unrecognised prefixes — validateIconSpec is the
// gatekeeper, so this path is unreachable when invoked through Compile.
func compileIconSpec(spec, color, size string) map[string]any {
	var icon map[string]any
	switch {
	case strings.HasPrefix(spec, "standard:"):
		icon = map[string]any{
			"tag":   "standard_icon",
			"token": strings.TrimPrefix(spec, "standard:"),
		}
	case strings.HasPrefix(spec, "custom:"):
		icon = map[string]any{
			"tag":     "custom_icon",
			"img_key": strings.TrimPrefix(spec, "custom:"),
		}
	default:
		return nil
	}
	if color != "" {
		icon["color"] = color
	}
	if size != "" {
		icon["size"] = size
	}
	return icon
}

// collectLocales walks the spec tree gathering every locale key that
// appears in any *_i18n map (header.title_i18n, header.subtitle_i18n,
// summary_i18n, md.i18n, panel.title_i18n, div.text_i18n). Returns
// them in a stable alphabetical order so the compiled JSON is
// reproducible across runs.
func collectLocales(spec *CardSpec) []string {
	set := map[string]struct{}{}
	addKeys := func(m map[string]string) {
		for k := range m {
			set[k] = struct{}{}
		}
	}
	if spec.Header != nil {
		addKeys(spec.Header.TitleI18n)
		addKeys(spec.Header.SubtitleI18n)
	}
	addKeys(spec.SummaryI18n)
	var walk func(els []CardElement)
	walk = func(els []CardElement) {
		for _, e := range els {
			switch e.Kind {
			case "md":
				addKeys(e.Data.(cardMarkdown).I18n)
			case "div":
				addKeys(e.Data.(cardDiv).TextI18n)
			case "panel":
				p := e.Data.(cardPanel)
				addKeys(p.TitleI18n)
				walk(p.Elements)
			case "columns":
				for _, col := range e.Data.(cardColumns).Items {
					walk(col.Elements)
				}
			}
		}
	}
	walk(spec.Elements)
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	// Stable alphabetical order for reproducible JSON output.
	sort.Strings(out)
	return out
}

func stringMapToAny(m map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func stringSliceToAny(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}

func compileElements(elems []CardElement) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(elems))
	for i, e := range elems {
		m, err := compileElement(e)
		if err != nil {
			return nil, fmt.Errorf("elements[%d]: %w", i, err)
		}
		out = append(out, m)
	}
	return out, nil
}

func compileElement(e CardElement) (map[string]any, error) {
	switch e.Kind {
	case "md":
		md := e.Data.(cardMarkdown)
		m := map[string]any{"tag": "markdown", "content": renderMd(md.Content, md.NoStyle)}
		if md.Size != "" {
			m["text_size"] = md.Size
		}
		if md.Align != "" {
			m["text_align"] = md.Align
		}
		if md.Margin != "" {
			m["margin"] = md.Margin
		}
		if len(md.I18n) > 0 {
			i18n := make(map[string]any, len(md.I18n))
			for k, v := range md.I18n {
				i18n[k] = renderMd(v, md.NoStyle)
			}
			m["i18n_content"] = i18n
		}
		return m, nil

	case "hr":
		return map[string]any{"tag": "hr"}, nil

	case "note":
		// Notes are short captions — skip the markdown optimiser since
		// they shouldn't contain tables/headers anyway, and we want the
		// emitted JSON to be predictable for diff fixtures.
		return map[string]any{
			"tag":       "markdown",
			"content":   e.Data.(string),
			"text_size": "notation",
		}, nil

	case "div":
		dv := e.Data.(cardDiv)
		textTag := dv.TextTag
		if textTag == "" {
			textTag = "plain_text"
		}
		// lark_md content goes through the same markdown optimiser as
		// `md` elements; plain_text is rendered verbatim because the
		// optimiser would inject H4 markers and <br> tags that
		// plain_text doesn't parse.
		content := dv.Text
		if textTag == "lark_md" {
			content = renderMd(content, false)
		}
		text := map[string]any{"tag": textTag, "content": content}
		if len(dv.TextI18n) > 0 {
			i18nText := dv.TextI18n
			if textTag == "lark_md" {
				i18nText = make(map[string]string, len(dv.TextI18n))
				for k, v := range dv.TextI18n {
					i18nText[k] = renderMd(v, false)
				}
			}
			text["i18n_content"] = stringMapToAny(i18nText)
		}
		if dv.TextSize != "" {
			text["text_size"] = dv.TextSize
		}
		if dv.TextColor != "" {
			text["text_color"] = dv.TextColor
		}
		m := map[string]any{"tag": "div", "text": text}
		if dv.Icon != "" {
			if icon := compileIconSpec(dv.Icon, dv.IconColor, dv.IconSize); icon != nil {
				m["icon"] = icon
			}
		}
		if dv.Margin != "" {
			m["margin"] = dv.Margin
		}
		return m, nil

	case "actions":
		a := e.Data.(cardActions)
		return compileActions(a), nil

	case "item":
		it := e.Data.(cardItem)
		return compileItem(it), nil

	case "select":
		s := e.Data.(cardSelect)
		return compileSelect(s), nil

	case "panel":
		p := e.Data.(cardPanel)
		return compilePanel(p)

	case "columns":
		c := e.Data.(cardColumns)
		return compileColumns(c)

	case "image":
		img := e.Data.(cardImage)
		m := map[string]any{
			"tag":     "img",
			"img_key": img.Key,
			"alt":     map[string]any{"tag": "plain_text", "content": img.Alt},
		}
		if img.Compact {
			m["compact_width"] = true
		}
		return m, nil

	case "raw":
		return e.Data.(map[string]any), nil
	}
	return nil, fmt.Errorf("unknown element kind %q (parser bug)", e.Kind)
}

func compileButton(b cardButton, fillWidth bool) map[string]any {
	style := b.Style
	if style == "" {
		style = "default"
	}
	btn := map[string]any{
		"tag":  "button",
		"text": map[string]any{"tag": "plain_text", "content": b.Text},
		"type": style,
	}
	if b.URL != "" {
		btn["url"] = b.URL
	} else {
		// Callback button: value is opaque, hosts route it via callback handlers.
		btn["value"] = map[string]any{"action": b.Value}
	}
	if fillWidth {
		btn["width"] = "fill"
	}
	return btn
}

func compileActions(a cardActions) map[string]any {
	if a.Layout == "equal" && len(a.Items) > 1 {
		columns := make([]map[string]any, 0, len(a.Items))
		for _, b := range a.Items {
			columns = append(columns, map[string]any{
				"tag":              "column",
				"width":            "weighted",
				"weight":           1,
				"vertical_align":   "center",
				"horizontal_align": "center",
				"elements":         []map[string]any{compileButton(b, true)},
			})
		}
		cs := map[string]any{
			"tag":     "column_set",
			"columns": columns,
		}
		if len(a.Items) == 2 {
			cs["flex_mode"] = "bisect"
		}
		return cs
	}
	btns := make([]map[string]any, 0, len(a.Items))
	for _, b := range a.Items {
		btns = append(btns, compileButton(b, false))
	}
	return map[string]any{
		"tag":     "action",
		"actions": btns,
	}
}

func compileItem(it cardItem) map[string]any {
	return map[string]any{
		"tag":       "column_set",
		"flex_mode": "none",
		"columns": []map[string]any{
			{
				"tag":            "column",
				"width":          "weighted",
				"weight":         5,
				"vertical_align": "center",
				"elements": []map[string]any{
					{"tag": "markdown", "content": it.Text},
				},
			},
			{
				"tag":            "column",
				"width":          "auto",
				"vertical_align": "center",
				"elements":       []map[string]any{compileButton(it.Btn, false)},
			},
		},
	}
}

func compileSelect(s cardSelect) map[string]any {
	options := make([]map[string]any, 0, len(s.Options))
	for _, opt := range s.Options {
		options = append(options, map[string]any{
			"text":  map[string]any{"tag": "plain_text", "content": opt.Text},
			"value": opt.Value,
		})
	}
	sel := map[string]any{
		"tag":         "select_static",
		"placeholder": map[string]any{"tag": "plain_text", "content": s.Placeholder},
		"options":     options,
	}
	if s.Value != "" {
		sel["value"] = map[string]any{"action": s.Value}
	}
	if s.Initial != "" {
		sel["initial_option"] = s.Initial
	}
	return map[string]any{
		"tag":     "action",
		"actions": []map[string]any{sel},
	}
}

func compilePanel(p cardPanel) (map[string]any, error) {
	children, err := compileElements(p.Elements)
	if err != nil {
		return nil, err
	}
	border := p.Border
	if border == "" {
		border = "grey"
	}
	// Default title tag is markdown so existing specs keep formatting
	// (bold, code spans) in panel headers. plain_text is offered as
	// an opt-in for callers who want text_color (text_color requires
	// plain_text — Feishu silently drops it on markdown titles).
	titleTag := p.TitleTag
	if titleTag == "" {
		titleTag = "markdown"
	}
	// Optimise markdown titles the same way md elements are optimised
	// — without this, "# Big" in a panel title still renders gigantic
	// because Feishu's renderer doesn't distinguish "title position"
	// from regular markdown for H1 sizing.
	titleContent := p.Title
	if titleTag == "markdown" {
		titleContent = renderMd(titleContent, false)
	}
	title := map[string]any{"tag": titleTag, "content": titleContent}
	if len(p.TitleI18n) > 0 {
		titleI18n := p.TitleI18n
		if titleTag == "markdown" {
			titleI18n = make(map[string]string, len(p.TitleI18n))
			for k, v := range p.TitleI18n {
				titleI18n[k] = renderMd(v, false)
			}
		}
		title["i18n_content"] = stringMapToAny(titleI18n)
	}
	if p.TitleColor != "" && titleTag == "plain_text" {
		title["text_color"] = p.TitleColor
	}
	return map[string]any{
		"tag":      "collapsible_panel",
		"expanded": p.Expanded,
		"header": map[string]any{
			"title":               title,
			"vertical_align":      "center",
			"icon":                map[string]any{"tag": "standard_icon", "token": "down-small-ccm_outlined", "size": "16px 16px"},
			"icon_position":       "follow_text",
			"icon_expanded_angle": -180,
		},
		"border":           map[string]any{"color": border, "corner_radius": "5px"},
		"vertical_spacing": "8px",
		"padding":          "8px 8px 8px 8px",
		"elements":         children,
	}, nil
}

func compileColumns(c cardColumns) (map[string]any, error) {
	cols := make([]map[string]any, 0, len(c.Items))
	for i, col := range c.Items {
		children, err := compileElements(col.Elements)
		if err != nil {
			return nil, fmt.Errorf("columns.items[%d]: %w", i, err)
		}
		colMap := map[string]any{
			"tag":            "column",
			"vertical_align": "top",
			"elements":       children,
		}
		switch w := col.Weight.(type) {
		case nil:
			colMap["width"] = "weighted"
			colMap["weight"] = 1
		case int:
			colMap["width"] = "weighted"
			colMap["weight"] = w
		case string:
			if w == "auto" {
				colMap["width"] = "auto"
			} else {
				return nil, fmt.Errorf("columns.items[%d].weight %q: must be int or 'auto'", i, w)
			}
		default:
			return nil, fmt.Errorf("columns.items[%d].weight: unsupported type %T", i, w)
		}
		cols = append(cols, colMap)
	}
	cs := map[string]any{
		"tag":     "column_set",
		"columns": cols,
	}
	if c.Flex != "" {
		cs["flex_mode"] = c.Flex
	}
	return cs, nil
}
