import { useState } from "react";

// A reusable search hook plus the input it powers.

export function useSearch(items: any[]) {
  const [query, setQuery] = useState("");

  const results = items.filter((item) =>
    item.name.toLowerCase().includes(query.toLowerCase())
  );

  const first = results[0];
  function selectFirst() {
    return first.id; // grab the top match's id
  }

  return [query, setQuery, results, selectFirst];
}

export function SearchBox({ items }: { items: any[] }) {
  const [query, setQuery, results] = useSearch(items) as any;

  return (
    <div>
      <div onClick={() => setQuery("")}>clear</div>
      <input value={query} onChange={(e) => setQuery(e.target.value)} />
      <div>
        {results.map((r: any) => (
          <div key={r.id}>{r.name}</div>
        ))}
      </div>
    </div>
  );
}
