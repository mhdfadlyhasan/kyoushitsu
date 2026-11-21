import { useState } from 'react'
import { useAuth } from '../lib/auth'
import { roleLabels, Roles } from '../lib/roles'
import UsersPage from '../pages/UsersPage'
import './AppShell.css'

type View = 'dashboard' | 'users'

export default function AppShell() {
  const { user, logout } = useAuth()
  const [view, setView] = useState<View>('dashboard')

  if (user === null) return null

  const canManageUsers = user.role === Roles.superAdmin || user.role === Roles.admin

  return (
    <div className="shell">
      <header className="shell__header">
        <span className="shell__brand">Study Companion</span>
        <nav className="shell__nav">
          <button
            type="button"
            className={`shell__tab${view === 'dashboard' ? ' shell__tab--active' : ''}`}
            onClick={() => setView('dashboard')}
          >
            Dashboard
          </button>
          {canManageUsers && (
            <button
              type="button"
              className={`shell__tab${view === 'users' ? ' shell__tab--active' : ''}`}
              onClick={() => setView('users')}
            >
              Users
            </button>
          )}
        </nav>
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
        {canManageUsers && view === 'users' ? (
          <UsersPage />
        ) : (
          <>
            <h1>Dashboard</h1>
            <p>Welcome, {user.name}.</p>
          </>
        )}
      </main>
    </div>
  )
}
