package terminal

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/raffepaffe/kload/internal/k8s"
)

const (
	// rootID is the ID assigned to the root container.
	rootID             = "root"
	showXMinValues     = 20
	showTotalMaxValues = 12
	title              = "Press Esc to quit (CPU in yellow, Memory in red)"
)

type dataSource interface {
	Fetch() ([]*k8s.Element, error)
	MaxColumns() int
}

func Draw(datasource dataSource) error {
	t, err := tcell.New()
	if err != nil {
		return fmt.Errorf("tcell new error: %w", err)
	}
	defer t.Close()

	allLineCharts := make(map[string]*linechart.LineChart)
	builder := grid.New()
	elements, err := datasource.Fetch()
	if err != nil {
		return fmt.Errorf("datasource.Fetch() => %w", err)
	}

	rowOfElements := make([]grid.Element, 0)

	maxElementsOnScreen := int(math.Min(float64(showTotalMaxValues), float64(len(elements))))
	columns, rows := chartFormat(datasource.MaxColumns(), maxElementsOnScreen)
	withPercentColumns := percent(1, columns)
	withPercentRows := percent(1, rows)

	for i := 0; i < maxElementsOnScreen; i++ {
		e := elements[i]
		lc, _ := newLineChartWidget()
		allLineCharts[e.Name] = lc
		title := fmt.Sprint(e.Name, " (", e.CPULimit, "Mi/", e.MemoryLimit, "MB)")
		rowOfElements = append(rowOfElements, grid.ColWidthPerc(withPercentColumns,
			grid.Widget(lc,
				container.ID(e.Name),
				container.Border(linestyle.Light),
				container.BorderTitle(title),
				container.BorderColor(cell.ColorCyan)),
		))
		if (i+1)%columns == 0 {
			row := grid.RowHeightPerc(withPercentRows, rowOfElements...)
			builder.Add(row)
			rowOfElements = nil
		}
	}

	if len(rowOfElements) > 0 {
		row := grid.RowHeightPerc(withPercentRows, rowOfElements...)
		builder.Add(row)
		rowOfElements = nil
	}

	gridOpts, err := builder.Build()
	if err != nil {
		return fmt.Errorf("builder.Build => %w", err)
	}

	rootContainer, err := container.New(t,
		container.ID(rootID),
		container.Border(linestyle.Light),
		container.BorderTitle(title))
	if err != nil {
		return fmt.Errorf("container.New => %w", err)
	}

	err = rootContainer.Update(rootID, gridOpts...)
	if err != nil {
		return fmt.Errorf("container.Update => %w", err)
	}

	// Esc to quit.
	exit := false
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			exit = true
		}
	}

	ctrl, err := termdash.NewController(t, rootContainer, termdash.KeyboardSubscriber(quitter))
	if err != nil {
		return fmt.Errorf("termdash.NewController => %w", err)
	}
	defer ctrl.Close()

	charts := make(map[string]*aChart)

	for i := 0; i < math.MaxInt; i++ {
		if err := ctrl.Redraw(); err != nil {
			return fmt.Errorf("ctrl.Redraw => %w", err)
		}
		if exit {
			break
		}

		elements, err = datasource.Fetch()
		if err != nil {
			return fmt.Errorf("datasource.Fetch() => %w", err)
		}
		if len(elements) == 0 {
			return fmt.Errorf("no elements found for command")
		}

		for _, element := range elements {
			lc, found := allLineCharts[element.Name]
			if found {
				chart, exist := charts[element.Name]
				if !exist {
					chart = &aChart{name: element.Name, xValues: make(map[int]string)}
					charts[element.Name] = chart
				}
				chart.cpu = append(chart.cpu, element.CPUPercent())
				chart.cpu = lastOfSlice(chart.cpu, 0, showXMinValues)
				chart.memory = append(chart.memory, element.MemoryPercent())
				chart.memory = lastOfSlice(chart.memory, 0, showXMinValues)
				chart.xValues[chart.step] = time.Now().Format("15:04:05")
				chart.xValues = lastOfMap(chart.xValues, showXMinValues)
				chart.step++

				err := lc.Series(chart.name+"cpu", chart.cpu,
					linechart.SeriesCellOpts(cell.FgColor(cell.ColorYellow)), linechart.SeriesXLabels(chart.xValues))
				if err != nil {
					return fmt.Errorf("lc.series error: %w", err)
				}
				err = lc.Series(chart.name+"memory", chart.memory, linechart.SeriesCellOpts(cell.FgColor(cell.ColorRed)))
				if err != nil {
					return fmt.Errorf("lc.series mem error: %w", err)
				}
			}
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

type aChart struct {
	name    string
	cpu     []float64
	memory  []float64
	step    int
	xValues map[int]string
}

// newLineChartWidget creates a new line chart.
func newLineChartWidget() (*linechart.LineChart, error) {
	lc, err := linechart.New(
		linechart.XAxisUnscaled(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.YAxisCustomScale(0, aHundred),
		linechart.YAxisFormattedValues(yAxisFormatter()),
	)
	if err != nil {
		return nil, fmt.Errorf("linechart.New => %w", err)
	}

	return lc, nil
}

// yAxisFormatter formats y-axis float64 to a string with no decimal values.
func yAxisFormatter() func(value float64) string {
	return func(value float64) string {
		return fmt.Sprintf("%.f", value)
	}
}

// lastOfSlice removes first from values in slice.
func lastOfSlice(inputs []float64, from, minSize int) []float64 {
	if len(inputs) < minSize {
		return inputs
	}

	copy(inputs[from:], inputs[from+1:])
	inputs = inputs[:len(inputs)-1]

	return inputs
}

// lastOfMap removes key with the lowest value from map.
func lastOfMap(inputs map[int]string, minSize int) map[int]string {
	if len(inputs) < minSize {
		return inputs
	}

	keys := make([]int, 0, len(inputs))
	for k := range inputs {
		keys = append(keys, k)
	}

	sort.Ints(keys)
	s := make(map[int]string, len(keys))
	for i := 0; i < len(keys)-1; i++ {
		k := keys[i+1]
		s[i] = inputs[k]
	}

	return s
}

// chartFormat returns the number of columns and row for the chart.
func chartFormat(maxColumns, numberOfElements int) (int, int) {
	if numberOfElements <= maxColumns {
		return numberOfElements, 1
	}

	rows := float64(numberOfElements) / float64(maxColumns)
	rows = math.Ceil(rows)
	rounded := int(math.Round(rows))

	return maxColumns, rounded
}

// percent calculates how many percent x is of y.
func percent(x, y int) int {
	p := float64(x) / float64(y)
	p *= aHundred

	return int(math.Round(p) - 1)
}

const aHundred = 100
