package main

import (
    "encoding/json"
    "fmt"
    "context"
    "os/exec"
    "time"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
//     "io/ioutil"
)

const uri = "mongodb://localhost"


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

        _, err = collection.InsertOne(context.TODO(), focus)
        if err != nil {
            panic(err)
        }
        
        time.Sleep(10 * time.Second)
    }

}
