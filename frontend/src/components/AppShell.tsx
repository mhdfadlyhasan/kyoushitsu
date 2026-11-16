import { useAuth } from '../lib/auth'
import type { Role } from '../lib/api'
import './AppShell.css'

const roleLabels: Record<Role, string> = {
  super_admin: 'Super admin',
  admin: 'Admin',
  teacher: 'Teacher',
  student: 'Student',
}

export default function AppShell() {
  const { user, logout } = useAuth()
  if (user === null) return null

  return (
    <div className="shell">
      <header className="shell__header">
        <span className="shell__brand">Study Companion</span>
        <div className="shell__account">
          <span className="shell__identity">
            <span className="shell__name">{user.name}</span>
            <span className="shell__role">{roleLabels[user.role]}</span>
          </span>
          <button type="button" className="shell__logout" onClick={logout}>
            Log out
          </button>
        </div>
      </header>

      <main className="shell__main">
        <h1>Dashboard</h1>
        <p>Welcome, {user.name}.</p>
      </main>
    </div>
  )
}
