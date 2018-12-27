package main

import (
    "image"
    "image/color"
    "fmt"
    "time"
    "strconv"
    "sort"
    "os"
    "github.com/globalsign/mgo"
    "github.com/globalsign/mgo/bson"
    "github.com/fogleman/gg"
    "math/rand"
)

type Task struct {
    StartTime time.Time
    EndTime time.Time
    TaskName string
}

type TaskType struct {
    Shares int
    TaskName string
    Color color.RGBA
}

// HLine draws a horizontal line
func HLine(img *image.RGBA, x1 int, x2 int, y int, col color.RGBA) {
    for ; x1 <= x2; x1++ {
        img.Set(x1, y, col)
    }
}

// VLine draws a veritcal line
func VLine(img *image.RGBA, x int, y1 int, y2 int, col color.RGBA) {
    for ; y1 <= y2; y1++ {
        img.Set(x, y1, col)
    }
}

// Rect draws a rectangle utilizing HLine() and VLine()
func Rect(img *image.RGBA, x1 int, y1 int, x2 int, y2 int, col color.RGBA) {
    HLine(img, x1, y1, x2, col)
    HLine(img, x1, y2, x2, col)
    VLine(img, x1, y1, y2, col)
    VLine(img, x2, y1, y2, col)
}


func get_range(start int64, end int64) ([]Task, []TaskType, map[string]int, int, error) {
    var err error
    session, err := mgo.Dial("localhost")

    if err != nil {
        fmt.Println(err)
        return nil, nil, nil, 0, err
    }
    defer session.Close()

    collection := session.DB("tracking").C("focus")
    var data []map[string]interface{}
    var tasks []Task
    var taskTypes []TaskType
    collection.Find(bson.M{ "timestamp" : bson.M{ "$gt" : start, "$lt" : end } }).All(&data)

    sort.Slice(data, func(i, j int) bool { return data[i]["timestamp"].(int64) < data[j]["timestamp"].(int64) })
    var oldclass string
    var curclass string
    var oldTs int64
    var curTs int64
    var startTask int64
    var endTask int64
    shares := make(map[string]int)
    sum_shares := 0
    temp_shares := 0
    for _, d := range data {
        if _, ok := d["class"]; !ok {
            continue
        }
        curclass = d["class"].(string)
        curTs = d["timestamp"].(int64)
        if oldclass == "" {
            oldclass = curclass
            startTask = d["timestamp"].(int64)
        }
        isPresent := false
        for _, t := range taskTypes {
            if t.TaskName == curclass {
                isPresent = true
            }
        }

        if !isPresent {
            taskTypes = append(taskTypes, TaskType{0, curclass, color.RGBA{0, 0, 0, 0}})
        }

        if oldclass != curclass || curTs - oldTs > 15 {
            endTask = d["timestamp"].(int64)
            if curTs - oldTs > 15 {
                endTask = oldTs
            }
            if endTask - startTask > 5 * 3600 {
                // more than 5 hours the same task is idle
                shares[oldclass] -= temp_shares
                oldclass = "idle"
                shares[oldclass] += temp_shares
            }
            tasks = append(tasks, Task{time.Unix(startTask, 0), time.Unix(endTask, 0), oldclass})
            startTask = curTs
            temp_shares = 0
        }
        oldclass = curclass
        oldTs = curTs
        shares[oldclass]++
        sum_shares++
        temp_shares++
    }
    endTask = data[len(data)-1]["timestamp"].(int64)
    tasks = append(tasks, Task{time.Unix(startTask, 0), time.Unix(endTask, 0), oldclass})

    return tasks, taskTypes, shares, sum_shares, nil
}


func main() {

    offset := 24 * 60 * 60
    if len(os.Args) > 1 {
        modifier := 60 * 60
        if len(os.Args) > 2 {
            switch os.Args[2] {
                case "s":
                    fallthrough
                case "S":
                    modifier = 1
                case "m":
                    fallthrough
                case "M":
                    modifier = 60
                case "d":
                    fallthrough
                case "D":
                    modifier = 24 * 60 * 60
                default:
            }
        }
        f, err := strconv.ParseFloat(os.Args[1], 64)
        if err != nil {
            fmt.Println(err)
            return
        }
        offset = int(f * float64(modifier))
    }

    target_timestamp := time.Now().Unix() - int64(offset)

    tasks, taskTypes, shares, sum_shares, err := get_range(target_timestamp, time.Now().Unix())
    if err != nil {
        fmt.Println(err)
        return
    }

    start_yearday := time.Unix(target_timestamp, 0).YearDay()
//    fmt.Println(taskTypes, tasks)

    colors := []color.RGBA{color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 255}, color.RGBA{0, 0, 255, 255}, color.RGBA{255, 255, 0, 255}, color.RGBA{255, 0, 255, 255}, color.RGBA{0, 255, 255, 255}, color.RGBA{255, 255, 255, 255}, color.RGBA{0, 0, 0, 255}, color.RGBA{85, 85, 85, 255}, color.RGBA{170, 170, 170, 255}, color.RGBA{128, 255, 0, 255}, color.RGBA{128, 0, 255, 255}, color.RGBA{255, 128, 0, 255}}
    //colornames := []string{"Red", "Green", "Blue", "Yellow", "Pink", "Light blue", "White", "Black", "Dark grey", "Light grey", "Lime", "Violet", "Orange"}
    fmt.Println(len(taskTypes), " vs ", len(colors))


    var sorted_shares []int
    var sorted_taskTypes []TaskType
    idle_shares := 0
    for _, v := range shares {
        sorted_shares = append(sorted_shares, v)
    }
    sort.Sort(sort.Reverse(sort.IntSlice(sorted_shares)))
    for _, i := range sorted_shares {
        for k, s := range shares {
            if s == i {
                for _, v := range taskTypes {
                    if v.TaskName == k {
                        v.Shares = s
                        sorted_taskTypes = append(sorted_taskTypes, v)
                        if k == "idle" {
                            idle_shares = s
                        }
                        break
                    }
                }
                break
            }
        }
    }

    var final_taskTypes []TaskType

    for i, ttype := range sorted_taskTypes {
        if len(colors) > i {
            ttype.Color = colors[i]
        } else {
            ttype.Color = color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255}
        }
        final_taskTypes = append(final_taskTypes, ttype)
    }

    fmt.Println(final_taskTypes)



    time_margin := 50
    bar_width := 120
    margin := 5
    graph_height := 1000
    legend_height := 15 * len(taskTypes) + 5
    date_margin := 30
    height := graph_height + legend_height + date_margin
    width := bar_width * ((offset / (24 * 3600)) + 1) + time_margin

    // mili-pixels per second = 1080 / (24 * 3600)
    mp_per_s := float64(1000 * graph_height) / float64(24 * 3600)
    fmt.Println("mp_per_s: ", mp_per_s)

    chart := image.NewRGBA(image.Rect(0, 0, width, height))
    for i := 0; i < height; i++ {
        HLine(chart, 0, width, i, color.RGBA{128, 128, 128, 255})
    }

    p_per_h := float64(graph_height) / float64(24)

    hour := 0
    bold := 1
    for i := float64(date_margin); i < float64(date_margin + graph_height); i += p_per_h {
        if hour == 12 {
            bold = 2
        } else {
            bold = 1
        }
        for j := int(i) - bold; j < int(i) + bold; j++ {
            HLine(chart, time_margin, width, j, color.RGBA{0, 0, 0, 255})
        }
        hour++
    }

    i := 1
    for _, ttype := range final_taskTypes {
        for j := 0; j < 10; j++ {
            HLine(chart, 5, 15, date_margin + graph_height + 5 + i + j, ttype.Color)
        }
        i += 15
    }
    days_in_year := 365
    for _, task := range tasks {
        h, m, s := task.StartTime.Clock()
        seconds := h * 3600 + m * 60 + s
        start_line := int(mp_per_s * float64(seconds)) / 1000
        h, m, s = task.EndTime.Clock()
        seconds = h * 3600 + m * 60 + s
        end_line := int(mp_per_s * float64(seconds)) / 1000
        day := task.StartTime.YearDay() - start_yearday
        if day < 0 {
            if task.StartTime.YearDay() == 1 {
                days_in_year = task.StartTime.Add(-24 * time.Hour).YearDay()
            }
            day += days_in_year
        }
        var ttype TaskType

        for _, tt := range final_taskTypes {
            if tt.TaskName == task.TaskName {
                ttype = tt
                break
            }
        }

        if end_line < start_line {
            // sanity check
            if (task.StartTime.YearDay() == task.EndTime.YearDay()) {
                fmt.Println("Impossible task found, bailing")
                continue
            }
            for i := start_line; i < graph_height; i++ {
                HLine(chart, time_margin + day * bar_width + margin, time_margin + (day + 1) * bar_width - margin, date_margin + i, ttype.Color)
            }
            day++
            for i := 0; i < end_line; i++ {
                HLine(chart, time_margin + day * bar_width + margin, time_margin + (day + 1) * bar_width - margin, date_margin + i, ttype.Color)
            }
        } else {
            for i := start_line; i < end_line; i++ {
                HLine(chart, time_margin + day * bar_width + margin, time_margin + (day + 1) * bar_width - margin, date_margin + i, ttype.Color)
            }
        }
    }

    dc := gg.NewContextForRGBA(chart)
    dc.SetRGBA(0, 0, 0, 1)
    i = 1
    h := 0
    for i := float64(date_margin); i < float64(date_margin + graph_height); i += p_per_h {
        str := fmt.Sprintf("%2v:00", h)
        dc.DrawString(str, 5, i + 3)
        h++
    }
    date := time.Unix(target_timestamp, 0)
    for i := float64(time_margin); i < float64(width); i += float64(bar_width) {
        y, m, d := date.Date()
        w := date.Weekday()
        line := fmt.Sprint(w.String()[0:3], " ", d, " ", m.String()[0:3], " ", y)
        dc.DrawString(line, i, float64(date_margin - 5))
        date = date.Add(24 * time.Hour)
    }

    for _, ttype := range final_taskTypes {
        line := fmt.Sprintf(": %v (%.2f%%)", ttype.TaskName, float64(ttype.Shares * 100) / float64(sum_shares))
        if ttype.TaskName != "idle" {
            line += fmt.Sprintf("(%.2f%%)", float64(ttype.Shares * 100) / float64(sum_shares - idle_shares))
        }
        dc.DrawString(line, 20, float64(date_margin + graph_height + (i * 15)))
        i++
    }
    dc.SavePNG("chart.png")
    return

}
