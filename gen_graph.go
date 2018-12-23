package main

import (
    "fmt"
    "os"
    "strconv"
    "github.com/globalsign/mgo"
    "github.com/globalsign/mgo/bson"
    "time"
    "io/ioutil"
    "bytes"
    "github.com/wcharczuk/go-chart"
)

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

    session, err := mgo.Dial("localhost")

    if err != nil {
        fmt.Println(err)
        return
    }

    defer session.Close()


    collection := session.DB("tracking").C("focus")
    var data []map[string]interface{}
    graph_data_title := make(map[string]int)
    graph_data_class := make(map[string]int)
    collection.Find(bson.M{ "timestamp" : bson.M{ "$gt" : target_timestamp } }).All(&data)
    //fmt.Println(data)
    for _, d := range data {
	if d["class"] == nil || d["title"] == nil {
		continue
	}
        graph_data_class[d["class"].(string)]++
        graph_data_title[d["title"].(string)]++
    }

    var class_vals []chart.Value
    var title_vals []chart.Value
    for k, gd := range graph_data_class {
        used_time := time.Duration(gd) * time.Second * 10
        percent := float64(gd * 100) / float64(len(data))
        class_vals = append(class_vals, chart.Value{Value: float64(gd), Label: k + " ("+ strconv.FormatFloat(percent, 'f', 2, 64) + "%/" + used_time.String() + ")"})
    }

    for k, gd := range graph_data_title {
        title_vals = append(title_vals, chart.Value{Value: float64(gd), Label: k})
        used_time := time.Duration(gd) * time.Second * 10
        percent := float64(gd * 100) / float64(len(data))
        title_vals = append(title_vals, chart.Value{Value: float64(gd), Label: k + " ("+ strconv.FormatFloat(percent, 'f', 2, 64) + "%/" + used_time.String() + ")"})
    }

    pie_class := chart.PieChart{
		Width:  1920,
		Height: 1080,
		Values: class_vals,
	}

	pie_title := chart.PieChart{
		Width:  1920,
		Height: 1080,
		Values: title_vals,
	}

    buffer := bytes.NewBuffer([]byte{})
    err = pie_class.Render(chart.PNG, buffer)
    if err != nil {
        fmt.Println(err)
        return
    }
    err = ioutil.WriteFile("class.png", buffer.Bytes(), 0644)
    if err != nil {
        fmt.Println(err)
        return
    }

    buffer = bytes.NewBuffer([]byte{})
    err = pie_title.Render(chart.PNG, buffer)
    if err != nil {
        fmt.Println(err)
        return
    }
    err = ioutil.WriteFile("title.png", buffer.Bytes(), 0644)
    if err != nil {
        fmt.Println(err)
        return
    }



}
