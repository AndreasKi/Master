package main

import (
	"math"
	"os"
	"strconv"
	"time"
)

func Output() {
	finished.Add(1)
	time := time.Now()
	os.Mkdir("logs/"+time.String(), 0777)
	html := CreateHTML(time)
	f, err := os.Create("logs/"+time.String()+"/Graphics.html")
	ErrorCheck(err, true)

	_, err = f.WriteString(html)
	ErrorCheck(err, true)
	f.Close()

	CreateDump(time)

	finished.Done()
	return
}

func CreateDump(time time.Time) {
	for index , d := range daemons {
		f, err := os.Create("logs/"+time.String()+"/dump_device_" + strconv.Itoa(index + 1) + ".txt")
		ErrorCheck(err, true)

		dump := "CPU:\n"
		for t, val:= range d.cpu_history {
			dump = dump + strconv.Itoa(d.start_time * interval + t * interval) + "\t" + strconv.FormatFloat(val, 'f', 1, 64) + "\n"
		}

		dump = dump + "\nMem:\n"
		for t, val:= range d.mem_history {
			dump = dump + strconv.Itoa(d.start_time * interval + t * interval) + "\t" + strconv.Itoa(val) + "\n"
		}

		_, err = f.WriteString(dump)
		ErrorCheck(err, true)
		f.Close()
	}
}

func CreateHTML(time time.Time) string {
	html := "<html><head><style>td {white-space: nowrap; overflow: hidden; text-overflow:ellipsis;}</style>" +
		"<title>Run Log; Stopped " + time.String() + "</title></head><body><center>"
	html = html + "<h1> Run Log; Stopped " + time.String() + " </h1><p>Started at time t = " + start.String() + "</p>"
	html = html + "<p>Logging ran for &#916; = " + time.Sub(start).String() + " with intervals of " + strconv.Itoa(interval) + "s</p>"
	html = html + "<h2> Overall Statistics </h2>"
	html = html + CreateCountHistogram("Number of Running \"Devices\"", num_daemons, 0)
	html = html + CreateCPUHistogram("Total CPU Usage", cpu_history, 0)
	html = html + CreateMemHistogram("Total Memory Usage", mem_history, 0)
	html = html + "</br>"
	for i, d := range daemons {
		t := d.start_time * interval
		html = html + "<h2> \"Device " + strconv.Itoa(i+1) + "\" with Daemon at " + d.ip + ":" + d.port + " </h2>"
		html = html + "<p>Found at t+" + strconv.Itoa(t) + "s</p>"
		html = html + CreateCPUHistogram("CPU Usage", d.cpu_history, t)
		html = html + CreateMemHistogram("Memory Usage", d.mem_history, t)
		html = html + "</br>"
	}
	html = html + "</center></body></html>"

	return html
}

func CreateCPUHistogram(title string, variable []float64, start_time int) string {
	max := FindFloatMax(variable)
	step := FindFloatStep(max)
	max = max + step
	hasVals := false
	histogram := "<h3>" + title + "</h3><table>"
	for y := max; y >= 0.0; y = SetPrecision(y-step, 2) {
		if !(y < step) && y != max {
			if step != 0.25 && step != 0.75 {
				histogram = histogram + "<tr><td style=\"font-size:5%;\">" + strconv.FormatFloat(y, 'f', 1, 64) + " &#37;</td>"
			} else {
				histogram = histogram + "<tr><td style=\"font-size:5%;\">" + strconv.FormatFloat(y, 'f', 2, 64) + " &#37;</td>"
			}
		} else {
			histogram = histogram + "<tr><td></td>"
		}
		for x, val := range variable {
			if val != 0.0 {
				hasVals = true
				if !(y < step) {
					open_tag := "<td>"
					close_tag := "</td>"
					content := ""
					if y == max {
						if val == -1.0 {
							content = "N/A"
						} else {
							content = strconv.FormatFloat(val, 'f', 1, 64) + " &#37;"
						}
					} else {
						if val >= y {
							open_tag = "<td bgcolor=\"blue\">"
						} else if val == -1.0 {
							open_tag = "<td bgcolor=\"red\">"
						}
					}
					histogram = histogram + open_tag + content + close_tag
				} else {
					histogram = histogram + "<td> t+" + strconv.Itoa(x*interval+start_time) + "s</td>"
				}
			}
		}
		histogram = histogram + "</tr>"
	}

	if hasVals && len(variable) > 0 {
		histogram = histogram + "</table>"
		return histogram
	} else if !hasVals {
		return "<h3>" + title + "</h3>No values were above 0.0 &#37;</h3></br>"
	} else {
		return "<h3>" + title + "</h3>Nothing to show, data set empty</h3></br>"
	}
}

func FindFloatMax(variable []float64) float64 {
	max := 0.00
	for _, val := range variable {
		if val > max {
			max = val
		}
	}
	return math.Ceil(max)
}

func FindFloatStep(max float64) float64 {
	step := 1.00
	if max > 100.00 {
		step = 5.0
	} else if max <= 100.00 && max > 60.00 {
		step = 2.0
	} else if max <= 60.00 && max > 30.00 {
		step = 1.0
	} else if max <= 30.00 && max > 10.00 {
		step = 0.75
	} else if max <= 10.00 && max > 5.00 {
		step = 0.50
	} else if max <= 5.00 && max > 1.00 {
		step = 0.25
	} else if max <= 1.00 {
		step = 0.10
	}
	return step
}

func SetPrecision(num float64, precision int) float64 {
	round := func(n float64) int {
		return int(n + math.Copysign(0.5, n))
	}
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func CreateCountHistogram(title string, variable []int, start_time int) string {
	PrevVal := func(index int) int {
		if index != 0 {
			return variable[index-1]
		} else {
			return -1
		}
	}

	max := FindCountMax(variable)
	hasVals := false
	histogram := "<h3>" + title + "</h3><table>"
	for y := max; y >= 0; y-- {
		if y != 0.0 && y != max {
			histogram = histogram + "<tr><td style=\"font-size:5%;\">" + strconv.Itoa(y) + "</td>"
		} else {
			histogram = histogram + "<tr><td></td>"
		}

		for x, val := range variable {
			if val != PrevVal(x) {
				hasVals = true
				if y != 0 {
					open_tag := "<td>"
					close_tag := "</td>"
					content := ""
					if y == max {
						if val == -1 {
							content = "N/A"
						} else {
							content = "<center>" + strconv.Itoa(val) + "</center>"
						}
					} else {
						if val >= y {
							open_tag = "<td bgcolor=\"blue\">"
						} else if val == -1 {
							open_tag = "<td bgcolor=\"red\">"
						}
					}
					histogram = histogram + open_tag + content + close_tag
				} else {
					histogram = histogram + "<td> t+" + strconv.Itoa(x*interval+start_time) + "s</td>"
				}
			}
		}
		histogram = histogram + "</tr>"
	}

	if hasVals {
		histogram = histogram + "</table>"
		return histogram
	} else {
		return "<h3>" + title + "</h3>Nothing to show, or no values were above 0"
	}
}

func FindCountMax(variable []int) int {
	max := 0
	for _, val := range variable {
		if val > max {
			max = val
		}
	}
	return max + 1
}

func CreateMemHistogram(title string, variable []int, start_time int) string {
	max := FindMemMax(variable)
	step := FindIntStep(max)
	max = max + step
	hasVals := false
	histogram := "<h3>" + title + "</h3><table>"
	for y := max; y >= 0; y = y - step {
		if !(y < step) && y != max {
			histogram = histogram + "<tr><td style=\"font-size:5%;\">" + strconv.Itoa(y) + "KB</td>"
		} else {
			histogram = histogram + "<tr><td></td>"
		}

		for x, val := range variable {
			if val != 0 {
				hasVals = true
				if !(y < step) {
					open_tag := "<td>"
					close_tag := "</td>"
					content := ""
					if y == max {
						if val == -1 {
							content = "N/A"
						} else {
							content = "<center>" + strconv.Itoa(val) + "KB</center>"
						}
					} else {
						if val >= y {
							open_tag = "<td bgcolor=\"blue\">"
						} else if val == -1 {
							open_tag = "<td bgcolor=\"red\">"
						}
					}
					histogram = histogram + open_tag + content + close_tag
				} else {
					histogram = histogram + "<td> t+" + strconv.Itoa(x*interval+start_time) + "s</td>"
				}
			}
		}
		histogram = histogram + "</tr>"
	}

	if hasVals {
		histogram = histogram + "</table>"
		return histogram
	} else {
		return "<h3>" + title + "</h3>Nothing to show, or no values were above 0"
	}
}

func FindIntStep(max int) int {
	step := max / 10
	return step
}
func FindMemMax(variable []int) int {
	max := 0
	for _, val := range variable {
		if val > max {
			max = val
		}
	}
	return max
}
