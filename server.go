package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

func serve_text(text string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) { fmt.Fprintf(w, text) }
}

func select_todo(w http.ResponseWriter, r *http.Request) {
	var id, err = strconv.Atoi(r.FormValue("id"))

	if err != nil {
		fmt.Fprint(w, build_todo_list(todo_list))
	}

	todo := todo_list[id]

	todo.Completed = true

	todo_list[id] = todo

	fmt.Fprint(w, build_todo_list(todo_list))
}

func fix_todo_list_ids(todos []todo_item) []todo_item {
	var copy []todo_item = todos

	for i, todo := range copy {
		todo.Todo_id = i
		copy[i] = todo
	}

	return copy
}

func delete_todo(w http.ResponseWriter, r *http.Request) {
	var id, err = strconv.Atoi(r.PathValue("id"))
	if err != nil {
		fmt.Println("Error Occurred: ")
		fmt.Println(err)
	}

	//
	todo_list = slices.Delete(todo_list, id, id+1) // This is not working...
	todo_list = fix_todo_list_ids(todo_list)

	fmt.Fprint(w, build_todo_list(todo_list))
}

func clear_todo_list(w http.ResponseWriter, _ *http.Request) {
	todo_list = []todo_item{}

	fmt.Fprint(w, build_todo_list(todo_list))
}

func save_todos_to_file(filename string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Saving to-dos to %v\n", filename)

		if !file_exists(filename) {
			os.Create(filename)
		}

		// Convert todo list to bytes and save it to the file
		bytes, err := json.Marshal(todo_list)
		if err != nil {
			fmt.Println(err)
			fmt.Println("There was an error saving data.")
			return
		}

		err = os.WriteFile(filename, bytes, 0644)
		if err != nil {
			fmt.Println("There was a error writing the file")
			fmt.Println(err)
		}
	}
}

func load_todos_from_file(filename string) {
	fmt.Println("Loading Todo List...")

	data, err := os.ReadFile(filename)

	if err != nil {
		fmt.Println("Error Reading file")
		fmt.Println(err)
		return
	}

	json.Unmarshal(data, &todo_list)
}

func deselect_todo(w http.ResponseWriter, r *http.Request) {
	var id, err = strconv.Atoi(r.FormValue("id"))
	if err != nil {
		fmt.Fprint(w, build_todo_list(todo_list))
	}

	todo := todo_list[id]

	todo.Completed = false

	todo_list[id] = todo

	fmt.Fprint(w, build_todo_list(todo_list))
}

func template_format(s string, replace_list [][]string) string {
	var final string = s

	for _, pair := range replace_list {
		final = strings.ReplaceAll(final, pair[0], pair[1])
	}

	return final
}

func file_exists(filename string) bool {
	_, err := os.Stat(filename)
	return !(errors.Is(err, os.ErrNotExist))
}

func build_todo_list(todos []todo_item) string {
	const todo_template string = `
    <div class="todo-item todo-TODOTYPE" id="todo-1">
      <p class="LINETHROUGH">TODOTEXT</p>
      <input type="hidden" hx-trigger="click from:next"  name="id" value="ID" hx-target="closest .todo-list" hx-swap="innerHTML" hx-post="ENDPOINT">
      <input ISCHECKED type="checkbox" name="id">
      <button hx-post="/delete-todo/ID" hx-target="closest .todo-list" hx-swap="innerHTML" value="ID">Delete</button>
    </div>
    `
	const form_template string = `
    <form hx-post="/new-todo" hx-target="closest #todo-list" hx-swap="innerHTML" hx-trigger="submit">
      <input type="text" name="todo-text" placeholder="Enter Todo Item">
      <input type="submit" name="submit" value="Create New Todo">
    </form>
    `

	var html string

	for _, todo := range todos {
		if !todo.Completed {
			replace_list := make([][]string, 0)
			replace_list = append(replace_list,
				[]string{"ISCHECKED", ""},
				[]string{"LINETHROUGH", ""},
				[]string{"TODOTEXT", todo.Todo_text},
				[]string{"ID", fmt.Sprint(todo.Todo_id)},
				[]string{"ENDPOINT", "/task-checkoff"},
				[]string{"TODOTYPE", fmt.Sprint(todo.Todo_id % 2)},
			)

			html += template_format(todo_template, replace_list)

		} else {
			replace_list := make([][]string, 0)
			replace_list = append(replace_list,
				[]string{"ISCHECKED", "checked"},
				[]string{"LINETHROUGH", "checked-off"},
				[]string{"TODOTEXT", todo.Todo_text},
				[]string{"ID", fmt.Sprint(todo.Todo_id)},
				[]string{"ENDPOINT", "/task-uncheckoff"},
				[]string{"TODOTYPE", fmt.Sprint(todo.Todo_id % 2)},
			)

			html += template_format(todo_template, replace_list)
		}
	}

	html += form_template
	return html
}

func create_todo(w http.ResponseWriter, r *http.Request) {
	var Todo_text string = r.FormValue("todo-text")

  todo_list = append(
    todo_list,
    todo_item{Todo_text, false, len(todo_list)},
    )

	fmt.Fprintf(w, build_todo_list(todo_list))
}

func print_todo_list(todos []todo_item) {
	for _, todo := range todos {
		fmt.Printf(todo.Todo_text)
		fmt.Printf(", ")
	}
	fmt.Println("")
}

func fetch_todos(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, build_todo_list(todo_list))
}

type todo_item struct {
	Todo_text string
	Completed bool
	Todo_id   int
}

func auto_save(filename string) {
	for {
		var old []todo_item = todo_list
		time.Sleep(time.Second * 10)

		if reflect.DeepEqual(old, todo_list) {
			continue
		}

		if !file_exists(filename) {
			os.Create(filename)
		}

		// Convert todo list to bytes and save it to the file
		bytes, err := json.Marshal(todo_list)
		if err != nil {
			fmt.Println(err)
			fmt.Println("There was an error saving data.")
			return
		}

		err = os.WriteFile(filename, bytes, 0644)
		if err != nil {
			fmt.Println("There was a error writing the file")
			fmt.Println(err)
		}
	}
}

var todo_list []todo_item = make([]todo_item, 0)

func main() {

	load_todos_from_file("data.json")

	go auto_save("data.json")

	mux := http.NewServeMux()

	mux.HandleFunc("/save-todos", save_todos_to_file("data.json"))
	mux.HandleFunc("/delete-todo/{id}", delete_todo)
	mux.HandleFunc("/fetch-todos", fetch_todos)
	mux.HandleFunc("/new-todo", create_todo)
	mux.HandleFunc("/task-checkoff", select_todo)
	mux.HandleFunc("/task-uncheckoff", deselect_todo)

	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./content"))))

	fmt.Println("Listening on port 3000...")

	http.ListenAndServe(":3000", mux)

}
