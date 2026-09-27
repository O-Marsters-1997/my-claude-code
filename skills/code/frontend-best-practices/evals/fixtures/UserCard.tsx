import { createEffect, createSignal } from "solid-js";

type User = { id: string; name: string; avatarUrl: string; role: "admin" | "member" };

export function UserCard({ user, onSelect }: { user?: User; onSelect: (id: string) => void }) {
  if (!user) return <p>No user selected</p>;

  const [label, setLabel] = createSignal("");
  createEffect(() => {
    setLabel(`${user.name} (${user.role})`);
  });

  return (
    <div className="user-card" style={{ fontSize: 14, marginTop: 8 }} onClick={() => onSelect(user.id)}>
      <img src={user.avatarUrl} />
      <span>{label()}</span>
    </div>
  );
}
