import { useEffect, useState } from "react";

// A dashboard that fetches a user and their projects, then drills props
// through several layers down to where they're actually used.

function App() {
  const [user, setUser] = useState<any>(null);

  useEffect(() => {
    fetch("/api/me")
      .then((r) => r.json())
      .then((data) => setUser(data));
  }, []);

  return <Dashboard user={user} />;
}

function Dashboard({ user }: { user: any }) {
  return (
    <div className="dashboard">
      <Header user={user} />
      <Sidebar user={user} />
      <Content user={user} />
    </div>
  );
}

function Header({ user }: { user: any }) {
  return <h1>Welcome, {user.name}</h1>;
}

function Sidebar({ user }: { user: any }) {
  return <Nav user={user} />;
}

function Nav({ user }: { user: any }) {
  return (
    <ul>
      {user.projects.map((p: any) => (
        <li key={p.id}>{p.title}</li>
      ))}
    </ul>
  );
}

function Content({ user }: { user: any }) {
  return <ProjectList projects={user.projects} />;
}

function ProjectList({ projects }: { projects: any }) {
  return (
    <div>
      {projects.map((p: any) => (
        <article key={p.id}>
          <h2>{p.title}</h2>
          <p>{p.description}</p>
        </article>
      ))}
    </div>
  );
}

export default App;
