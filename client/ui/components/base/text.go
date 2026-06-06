package base

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

const (
	H1Size = 22
	H2Size = 18
	H3Size = 16
)

type TextStyle struct {
	Size  float32
	Color color.Color
	Bold  bool
}

// Outgoing API methods
func H1(text string, style TextStyle) *TextWidget {
	return NewText(text, resolveTextSize(style.Size, H1Size), style.Color, fyne.TextStyle{Bold: style.Bold})
}

func H1Binding(text binding.String, style TextStyle) *TextWidget {
	return NewTextWithBinding(text, resolveTextSize(style.Size, H1Size), style.Color, fyne.TextStyle{Bold: style.Bold})
}

func H2(text string, style TextStyle) *TextWidget {
	return NewText(text, resolveTextSize(style.Size, H2Size), style.Color, fyne.TextStyle{Bold: style.Bold})
}

func H2Binding(text binding.String, style TextStyle) *TextWidget {
	return NewTextWithBinding(text, resolveTextSize(style.Size, H2Size), style.Color, fyne.TextStyle{Bold: style.Bold})
}

func H3(text string, style TextStyle) *TextWidget {
	return NewText(text, resolveTextSize(style.Size, H3Size), style.Color, fyne.TextStyle{Bold: style.Bold})
}

// General text component with customizable style
func Text(text string, style TextStyle) *TextWidget {
	return NewText(text, resolveTextSize(style.Size, 14), style.Color, fyne.TextStyle{Bold: style.Bold})
}

// Internal widget implementation
type TextWidget struct {
	widget.BaseWidget

	staticText string
	boundText  binding.String
	listener   binding.DataListener

	Color      color.Color
	TextSize   float32
	TextStyle  fyne.TextStyle
	Alignment  fyne.TextAlign
	FontSource fyne.Resource
}

func NewText(value string, size float32, col color.Color, style fyne.TextStyle) *TextWidget {
	t := &TextWidget{
		staticText: value,
		Color:      col,
		TextSize:   size,
		TextStyle:  style,
	}
	t.ExtendBaseWidget(t)
	return t
}

func NewTextWithBinding(value binding.String, size float32, col color.Color, style fyne.TextStyle) *TextWidget {
	t := &TextWidget{
		boundText: value,
		Color:     col,
		TextSize:  size,
		TextStyle: style,
	}
	t.ExtendBaseWidget(t)

	t.listener = binding.NewDataListener(func() {
		t.Refresh()
	})
	value.AddListener(t.listener)

	return t
}

func (t *TextWidget) SetText(value string) {
	if t.boundText != nil {
		_ = t.boundText.Set(value)
		return
	}

	t.staticText = value
	t.Refresh()
}

func (t *TextWidget) CreateRenderer() fyne.WidgetRenderer {
	text := canvas.NewText("", t.Color)
	text.TextSize = t.TextSize
	text.TextStyle = t.TextStyle
	text.Alignment = t.Alignment
	text.FontSource = t.FontSource

	r := &textRenderer{
		widget: t,
		text:   text,
	}
	r.Refresh()
	return r
}

type textRenderer struct {
	widget *TextWidget
	text   *canvas.Text
}

func (r *textRenderer) Layout(size fyne.Size) {
	r.text.Resize(size)
}

func (r *textRenderer) MinSize() fyne.Size {
	return r.text.MinSize()
}

func (r *textRenderer) Refresh() {
	if r.widget.boundText != nil {
		value, err := r.widget.boundText.Get()
		if err == nil {
			r.text.Text = value
		}
	} else {
		r.text.Text = r.widget.staticText
	}

	r.text.Color = r.widget.Color
	r.text.TextSize = r.widget.TextSize
	r.text.TextStyle = r.widget.TextStyle
	r.text.Alignment = r.widget.Alignment
	r.text.FontSource = r.widget.FontSource
	r.text.Refresh()
}

func (r *textRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text}
}

func (r *textRenderer) Destroy() {
	if r.widget.boundText != nil && r.widget.listener != nil {
		r.widget.boundText.RemoveListener(r.widget.listener)
	}
}

func resolveTextSize(size float32, fallback float32) float32 {
	if size <= 0 {
		return fallback
	}

	return size
}
