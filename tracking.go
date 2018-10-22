package main

import (
    "encoding/json"
    "fmt"
    "os/exec"
    "time"
    "github.com/globalsign/mgo"
//     "io/ioutil"
)


func get_focused_window(indent string, node map[string]interface{}) map[string]interface{} {
    if val, ok := node["nodes"]; ok {
        for _, node := range val.([]interface{}) {
            res := get_focused_window(indent + " ", node.(map[string]interface{}))
            if res != nil {
                return res
            }
        //return get_focused_window(indent + " ", val.(map[string]interface{}))
        }
    }
    if node["focused"].(bool) {
        if node["window_properties"] != nil {
            return node["window_properties"].(map[string]interface{})
        }
    }

    return nil

}

func main() {

    session, err := mgo.Dial("localhost")

    if err != nil {
        fmt.Println(err)
        return
    }

    defer session.Close()

    fmt.Println(session)

    collection := session.DB("tracking").C("focus")

    fmt.Println(collection)

    for {
        jsonStr, err := exec.Command("i3-msg", "-t", "get_tree").Output()
        if err != nil {
            fmt.Println(err)
            continue
        }

        data := make(map[string]interface{})
        err = json.Unmarshal([]byte(jsonStr), &data)

        if err != nil {
            fmt.Println(err)
            continue
        }

        focus := get_focused_window(" ", data)
//      fmt.Println("")
        if focus == nil {
            focus = make(map[string]interface{})
            focus["class"] = "idle"
            focus["title"] = "idle"
        }
        focus["timestamp"] = time.Now().Unix()
//              fmt.Println(focus)

        collection.Insert(focus)
        time.Sleep(10 * time.Second)
    }

}
