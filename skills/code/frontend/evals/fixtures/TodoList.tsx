import { createStore } from "solid-js/store";
import { For } from "solid-js";

type Todo = { id: number; title: string; done: boolean };

export function TodoList() {
  const [todos, setTodos] = createStore<Todo[]>([
    { id: 1, title: "Buy milk", done: false },
    { id: 2, title: "Walk dog", done: false },
  ]);

  function rename(todo: Todo, title: string) {
    // edit the item in place
    todo.title = title;
    setTodos([...todos]);
  }

  return (
    <ul>
      <For each={todos}>
        {(todo) => (
          <li>
            <input
              value={todo.title}
              onInput={(e) => rename(todo, e.currentTarget.value)}
            />
            {todo.done ? " ✓" : ""}
          </li>
        )}
      </For>
    </ul>
  );
}
