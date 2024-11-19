package main

import (
    "fmt"
    "html/template"
    "net/http"
    "time"
    "strconv"
    "sort"
    "strings"
)

type Task struct {
    ID          int
    Title       string
    Deadline    time.Time
    Done        bool
    Priority    string
    CreatedAt   time.Time
    IsOverdue   bool
}

type Statistics struct {
    CompletedRate    float64
    OverdueRate     float64
    PendingRate     float64
    TotalTasks      int
    CompletedTasks  int
    OverdueTasks    int
    PendingTasks    int
}

var (
    tasks []Task
    nextID = 1
)

func main() {
    go checkOverdueTasks()

    http.HandleFunc("/", handleHome)
    http.HandleFunc("/add", handleAdd)
    http.HandleFunc("/toggle/", handleToggle)
    http.HandleFunc("/delete/", handleDelete)
    http.HandleFunc("/sort/", handleSort)

    fmt.Println("Сервер запущен на http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}

func checkOverdueTasks() {
    for {
        now := time.Now()
        for i := range tasks {
            if !tasks[i].Done && tasks[i].Deadline.Before(now) {
                tasks[i].IsOverdue = true
            }
        }
        time.Sleep(1 * time.Minute)
    }
}

func calculateStatistics() Statistics {
    now := time.Now()
    var stats Statistics
    stats.TotalTasks = len(tasks)

    for _, task := range tasks {
        if task.Done {
            stats.CompletedTasks++
        } else if task.Deadline.Before(now) {
            stats.OverdueTasks++
        } else {
            stats.PendingTasks++
        }
    }

    if stats.TotalTasks > 0 {
        stats.CompletedRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
        stats.OverdueRate = float64(stats.OverdueTasks) / float64(stats.TotalTasks) * 100
        stats.PendingRate = float64(stats.PendingTasks) / float64(stats.TotalTasks) * 100
    }

    return stats
}

func handleHome(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/index.html"))
    
    stats := calculateStatistics()
    data := struct {
        Tasks []Task
        Stats Statistics
    }{
        Tasks: tasks,
        Stats: stats,
    }

    tmpl.Execute(w, data)
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    title := r.FormValue("title")
    deadline := r.FormValue("deadline")
    priority := r.FormValue("priority")

    deadlineTime, _ := time.Parse("2006-01-02T15:04", deadline)

    task := Task{
        ID:        nextID,
        Title:     title,
        Deadline:  deadlineTime,
        Priority:  priority,
        CreatedAt: time.Now(),
        IsOverdue: deadlineTime.Before(time.Now()),
    }

    tasks = append(tasks, task)
    nextID++

    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleToggle(w http.ResponseWriter, r *http.Request) {
    idStr := strings.TrimPrefix(r.URL.Path, "/toggle/")
    id, _ := strconv.Atoi(idStr)

    for i := range tasks {
        if tasks[i].ID == id {
            tasks[i].Done = !tasks[i].Done
            break
        }
    }

    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
    idStr := strings.TrimPrefix(r.URL.Path, "/delete/")
    id, _ := strconv.Atoi(idStr)

    for i := range tasks {
        if tasks[i].ID == id {
            tasks = append(tasks[:i], tasks[i+1:]...)
            break
        }
    }

    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleSort(w http.ResponseWriter, r *http.Request) {
    sortBy := strings.TrimPrefix(r.URL.Path, "/sort/")
    
    switch sortBy {
    case "deadline":
        sort.Slice(tasks, func(i, j int) bool {
            return tasks[i].Deadline.Before(tasks[j].Deadline)
        })
    case "priority":
        sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Priority > tasks[j].Priority
        })
    case "status":
        sort.Slice(tasks, func(i, j int) bool {
            return tasks[i].Done && !tasks[j].Done
        })
    }

    http.Redirect(w, r, "/", http.StatusSeeOther)
}

