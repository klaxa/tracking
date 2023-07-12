package main

import (
    "image"
    "image/color"
    "context"
    "fmt"
    "time"
    "strconv"
    "sort"
    "os"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "github.com/fogleman/gg"
    "math/rand"
)

const uri = "mongodb://localhost"

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

func fmtDuration(d time.Duration) string {
    d = d.Round(time.Minute)
    h := d / time.Hour
    d -= h * time.Hour
    m := d / time.Minute
    return fmt.Sprintf("%02d:%02d", h, m)
}


func get_range(start int64, end int64) ([]Task, []TaskType, map[string]int, int, []int, error) {
    var err error
    //serverAPI := options.ServerAPI(options.ServerAPIVersion1)
    opts := options.Client().ApplyURI(uri) //.SetServerAPIOptions(serverAPI)
    // Create a new client and connect to the server
    client, err := mongo.Connect(context.TODO(), opts)
    if err != nil {
        panic(err)
    }
    defer func() {
        if err = client.Disconnect(context.TODO()); err != nil {
            panic(err)
        }
    }()
    // Send a ping to confirm a successful connection
    var result bson.M
    if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
        panic(err)
    }
    
    fmt.Println(client)
    
    collection := client.Database("tracking").Collection("focus")
    
    fmt.Println(collection)
    var data []map[string]interface{}
    var tasks []Task
    var taskTypes []TaskType
    cursor, err := collection.Find(context.TODO(), bson.M{ "timestamp" : bson.M{ "$gt" : start, "$lt" : end } })
    if err != nil {
        panic(err)
    }
    err = cursor.All(context.TODO(), &data)
    if err != nil {
        panic(err)
    }
    sort.Slice(data, func(i, j int) bool { return data[i]["timestamp"].(int64) < data[j]["timestamp"].(int64) })
    var daily_shares []int
    start_day := time.Unix(start, 0)
    first_day := time.Unix(data[0]["timestamp"].(int64), 0)
    for start_day.Day() != first_day.Day() {
        daily_shares = append(daily_shares, 0)
        start_day = start_day.Add(24 * time.Hour)
    }
    current_day := 0
    current_shares := 0
    for i := 1; i < len(data); i++ {
        current_shares++
        previous_time := time.Unix(data[i-1]["timestamp"].(int64), 0)
        current_time := time.Unix(data[i]["timestamp"].(int64), 0)
        if current_time.Day() != previous_time.Day() {
            next_day := previous_time.Add(24 * time.Hour)
            fmt.Println(next_day.Day(), previous_time.Day())
            daily_shares = append(daily_shares, current_shares)
            current_day++
            current_shares = 0
            for next_day.Day() != current_time.Day() {
                fmt.Println("skip: ", next_day.Day(), previous_time.Day())
                daily_shares = append(daily_shares, 0)
                next_day = next_day.Add(24 * time.Hour)
            }
        }
    }
    daily_shares = append(daily_shares, current_shares)


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

    return tasks, taskTypes, shares, sum_shares, daily_shares, nil
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

    tasks, taskTypes, shares, sum_shares, daily_shares, err := get_range(target_timestamp, time.Now().Unix())
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
    daily_time_margin := 70
    height := graph_height + legend_height + date_margin + daily_time_margin
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
            HLine(chart, 5, 15, date_margin + graph_height + daily_time_margin + 5 + i + j, ttype.Color)
        }
        i += 15
    }
    days_in_year := 365
    for _, task := range tasks {
        // dc := gg.NewContextForRGBA(chart)
        // dc.SavePNG("chart." + strconv.Itoa(png_count) + ".png")
        // png_count++
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
    // print time on the left
    i = 1
    h := 0
    for i := float64(date_margin); i < float64(date_margin + graph_height); i += p_per_h {
        str := fmt.Sprintf("%2v:00", h)
        dc.DrawString(str, 5, i + 3)
        h++
    }

    // print date at the top and daily/monthly usage at the bottom
    date := time.Unix(target_timestamp, 0)
    current_day := 0
    current_month_duration := 0 * time.Second
    current_month_duration_round := 0 * time.Second
    for i := float64(time_margin); i < float64(width); i += float64(bar_width) {
        y, m, d := date.Date()
        w := date.Weekday()
        line := fmt.Sprint(w.String()[0:3], " ", d, " ", m.String()[0:3], " ", y)
        dc.DrawString(line, i, float64(date_margin - 5))
        date = date.Add(24 * time.Hour)
        _, next_m, _ := date.Date()
        day_duration, _ := time.ParseDuration(fmt.Sprint(daily_shares[current_day]) + "0s")
        current_month_duration += day_duration
        line = fmtDuration(day_duration)
        dc.DrawString(line, i + 30, float64(date_margin + graph_height + 15))
        day_duration_round := (day_duration + 59 * time.Minute).Truncate(time.Hour)
        current_month_duration_round += day_duration_round
        line = fmtDuration(day_duration_round)
        dc.DrawString(line, i + 30, float64(date_margin + graph_height + 30))
        current_day++
        if next_m != m {
            line = fmtDuration(current_month_duration)
            dc.DrawString(line, i + 30, float64(date_margin + graph_height + 45))
            line = fmtDuration(current_month_duration_round)
            dc.DrawString(line, i + 30, float64(date_margin + graph_height + 60))
            current_month_duration = 0 * time.Second
            current_month_duration_round = 0 * time.Second
        }
    }
    line := fmtDuration(current_month_duration)
    dc.DrawString(line, float64(width - bar_width + 30), float64(date_margin + graph_height + 45))
    line = fmtDuration(current_month_duration_round)
    dc.DrawString(line, float64(width - bar_width + 30), float64(date_margin + graph_height + 60))


    for _, ttype := range final_taskTypes {
        line := fmt.Sprintf(": %v (%v)(%.2f%%)", ttype.TaskName, 10 * time.Second * time.Duration(ttype.Shares), float64(ttype.Shares * 100) / float64(sum_shares))
        if ttype.TaskName != "idle" {
            line += fmt.Sprintf("(%.2f%%)", float64(ttype.Shares * 100) / float64(sum_shares - idle_shares))
        } else {
            line += fmt.Sprintf(" Total time: %v", 10 * time.Second * time.Duration(sum_shares))
        }
        dc.DrawString(line, 20, float64(date_margin + graph_height + daily_time_margin + (i * 15)))
        i++
    }
    dc.SavePNG("chart.png")
    return

}
