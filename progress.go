package progresscli

import (
    "os"
    "io"
    "fmt"
    "unicode/utf8"
    "math"
    "regexp"

    "github.com/nathan-fiscaletti/consolesize-go"
)

// Style represents the style that can be applied to a progress bar.
type Style struct {
    // The open and close characters are the characters on either end
    // of the progress bar. They can be used to encapsulate the
    // progress bar itself.
    OpenChar        string
    CloseChar       string

    // The done character is the character used to represent a
    // completed section of the progress bar.
    DoneChar        string

    // The not-done character is the character used to represent a
    // section of the progress bar that has not yet been completed.
    NotDoneChar     string

    // The in-progress character is the character used to represent a
    // section of the progress bar that is currently in progress.
    InProgressChar  string

    // The percentage color is the text that can be placed immediately
    // before the percentage print out and is most commonly used for
    // ANSI escape sequences to change the color of the text.
    PercentageColor string
}

// ProgressBar represents an instance of a Progress Bar. You should
// initialize a new progress-bar using the New() or NewWithStyle()
// functions.
type ProgressBar struct {
    style                 Style
    max                   float64
    showPercentage        bool
    showPercentageDecimal bool
    label                 string
    showLabel             bool
    writer                io.Writer
    value                 float64
    maxWidth              int
    useCustomMaxWidth     bool
    finished              bool
    visible               bool
}

// SetLabel sets the label for the progress bar. The label will be
// displayed on the left side of the progress bar.
func (pb *ProgressBar) SetLabel(label string) {
    pb.label = label
    pb.showLabel = strLen(label) > 0
    if pb.visible {
        pb.Increment(0)
    }
}

// SetShowPercentage will tell the progress bar to either display the
// current percentage or not to display it.
func (pb *ProgressBar) SetShowPercentage(show bool) {
    pb.showPercentage = show
    if pb.visible {
        pb.Increment(0)
    }
}

// SetShowPercentageDecimal will tell the progress bar to display the
// percentage with two character decimal precision. When called, this
// function will automatically force the percentage to be displayed,
// so it is not required that you also call SetShowPercentage(true).
func (pb *ProgressBar) SetShowPercentageDecimal(show bool) {
    if show {
        pb.showPercentage = true
    }

    pb.showPercentageDecimal = show
    if pb.visible {
        pb.Increment(0)
    }
}

// SetMax will set the maximum value for the progress bar. The default
// maximum value is 100.
func (pb *ProgressBar) SetMax(max float64) {
    pb.max = max
    if pb.visible {
        pb.Increment(0)
    }
}

// GetMax will retrieve the current max value for the progress bar.
func (pb *ProgressBar) GetMax() float64 {
    return pb.max
}

// SetMaxWidth will set the maximum width for the progress bar in 
// columns. The default value is the current width of the console.
func (pb *ProgressBar) SetMaxWidth(maxWidth int) {
    pb.maxWidth = maxWidth
    pb.useCustomMaxWidth = true
    if pb.visible {
        pb.Increment(0)
    }
}

// UseFullWidth will set the progress bar to use the current width in
// columns of the open console window. This is the default setting.
func (pb *ProgressBar) UseFullWidth() {
    pb.maxWidth = 0
    pb.useCustomMaxWidth = false
    if pb.visible {
        pb.Increment(0)
    }
}

// GetMaxWidth will retrieve the current maximum width of the
// progress bar in columns. If no custom maximum width has been set,
// the current width of the open console window will be returned.
func (pb *ProgressBar) GetMaxWidth() int {
    if pb.useCustomMaxWidth {
        return pb.maxWidth
    }

    cols, _ := consolesize.GetConsoleSize()
    return cols
}

// GetValue will retrieve the current value of the progress bar.
func (pb *ProgressBar) GetValue() float64 {
    return pb.value
}

// SetValue will set the current value of the progress bar.
func (pb *ProgressBar) SetValue(value float64) {
    pb.value = value
    if pb.visible {
        pb.Increment(0)
    }
}

// Show will show the progress bar in STDOUT.
func (pb *ProgressBar) Show() {
    pb.ShowIn(os.Stdout)
}

// ShowIn will show the progress bar in the specified io.Writer
func (pb *ProgressBar) ShowIn(w io.Writer) {
    pb.visible = true
    pb.writer = w
    pb.finished = false
    pb.value = 0
    pb.Increment(0)
}

// Increment will increment the progress bar by the specified count.
// The value of the progress bar will be constrained to 0-max where
// max is the current max value for the progress bar.
func (pb *ProgressBar) Increment(count float64) {
    if pb.finished || !pb.visible {
        return
    }

    pb.value += count
    if pb.value > pb.max {
        pb.value = pb.max
    }

    if pb.value < 0 {
        pb.value = 0
    }

    var output                   string
    var percent                  float64
    var labelLength              int
    var labelSpacerLength        int
    var percentLabel             string
    var percentLabelLength       int
    var percentLabelSpacerLength int

    var progressBarAvailableLength int
    var progressBarMinimumLength   int
    var labelsLength               int

    percent = (pb.value / pb.max) * 100.0;
    if !pb.showPercentageDecimal {
        percent = math.Trunc(percent)
    }

    if pb.showLabel {
        labelLength = strLen(pb.label)
        labelSpacerLength = 1
    }

    if pb.showPercentage {
        if pb.showPercentageDecimal {
            percentLabel = fmt.Sprintf("%.2f%%", percent)
            percentLabelLength = strLen(fmt.Sprintf("%.2f%%", 100.0))
        } else {
            percentLabel = fmt.Sprintf("%.0f%%", percent)
            percentLabelLength = strLen(fmt.Sprintf("%.0f%%", 100.0))
        }

        percentLabelSpacerLength = 1
    }

    if pb.showPercentage {
        labelsLength += percentLabelLength + percentLabelSpacerLength
    }

    if pb.showLabel {
        labelsLength += labelLength + labelSpacerLength
    }

    progressBarMinimumLength = strLen(pb.style.DoneChar) + 
                               strLen(pb.style.NotDoneChar) + 
                               strLen(pb.style.InProgressChar)
    cols, _ := consolesize.GetConsoleSize()
    if pb.useCustomMaxWidth { 
        progressBarAvailableLength = pb.maxWidth - 
                                     labelsLength - 
                                     strLen(pb.style.CloseChar) - 
                                     strLen(pb.style.OpenChar)
    } else {
        progressBarAvailableLength = cols - 
                                     labelsLength - 
                                     strLen(pb.style.CloseChar) - 
                                     strLen(pb.style.OpenChar)
    }

    // Clear the line before writing to it
    output += "\r"
    for i := 0; i<cols; i++ {
        output += " "
    }
    output += "\r"

    if progressBarAvailableLength < progressBarMinimumLength {
        if pb.showLabel && pb.showPercentage {
            output += fmt.Sprintf("%s %s", pb.label, percentLabel)
        } else if pb.showPercentage {
            output += fmt.Sprintf("%s", percentLabel)
        } else {
            output += fmt.Sprintf("%s", "Loading...")
        }
    } else {
        if pb.showLabel {
            output += fmt.Sprintf("%s ", pb.label)
        }

        output += fmt.Sprintf("%s", pb.style.OpenChar)

        var progressFillSize int
        progressFillSize = progressBarAvailableLength - 
                           strLen(pb.style.InProgressChar)
        filledBarLength := int(math.Trunc((percent / 100) * 
                               float64(progressFillSize)))

        if filledBarLength > 0 {
            for i := 0; i < filledBarLength; i++ {
                output += fmt.Sprintf("%s", pb.style.DoneChar)
            }
        }

        if strLen(pb.style.InProgressChar) > 0 {
            if percent < 100 {
                output += fmt.Sprintf("%s", pb.style.InProgressChar)
            } else {
                output += fmt.Sprintf("%s", pb.style.DoneChar)
            }
        }

        for j := 0; j < progressBarAvailableLength -
                        filledBarLength -
                        strLen(pb.style.InProgressChar); j++ {
            output += fmt.Sprintf("%s", pb.style.NotDoneChar)
        }

        if strLen(pb.style.CloseChar) > 0 {
            output += fmt.Sprintf("%s", pb.style.CloseChar)
        }

        if pb.showPercentage {
            output += fmt.Sprintf(
                " %s%4s", pb.style.PercentageColor, percentLabel)
        }
    }

    if percent >= 100 {
        pb.finished = true
        fmt.Fprintf(pb.writer, "%s\n", output)
    } else {
        fmt.Fprintf(pb.writer, "%s", output)
    }
}

// New will create a new progress bar using the default style.
func New() *ProgressBar {
    return NewWithStyle(DefaultStyle())
}

// NewWithStyle will create a new progress bar using the specified
// style object.
func NewWithStyle(style Style) *ProgressBar {
    return &ProgressBar{
        style: style,
        max: 100.0,
        showLabel: false,
        showPercentage: true,
    }
}

// DefaultStyle will retrieve the default Style for progress bars.
func DefaultStyle() Style {
    return Style {
        OpenChar: "",
        CloseChar: "",
        DoneChar: "\033[1;32m█\033[0m",
        NotDoneChar: "\033[1;37m░\033[0m",
        InProgressChar: "\033[1;37m░\033[0m",
    }
}

// DefaultStyleNoColor will retrieve the default Style for progress
// bars without any ANSI color escape sequences.
func DefaultStyleNoColor() Style {
    return Style {
        OpenChar: "",
        CloseChar: "",
        DoneChar: "█",
        NotDoneChar: "░",
        InProgressChar: "░",
    }
}

// LineStyle will retrieve a line type Style for progress bars.
func LineStyle() Style {
    return Style {
        OpenChar: "\033[1;37m╠\033[0m",
        CloseChar: "\033[1;37m╣\033[0m",
        DoneChar: "\033[1;32m═\033[0m",
        NotDoneChar: "\033[1;37m─\033[0m",
        InProgressChar: "\033[1;37m─\033[0m",
    }
}

// LineStyleNoColor will retrieve a line type Style for progress bars
// without any ANSI color escape sequences.
func LineStyleNoColor() Style {
    return Style {
        OpenChar: "╠",
        CloseChar: "╣",
        DoneChar: "═",
        NotDoneChar: "─",
        InProgressChar: "─",
    }
}

const ansi  = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
var ansi_re = regexp.MustCompile(ansi)
func strLen(s string) int {
    return utf8.RuneCountInString(ansi_re.ReplaceAllString(s, ""))
}