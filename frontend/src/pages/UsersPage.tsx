import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { apiRequest } from '../lib/api'
import type { User } from '../lib/api'
import { useAuth } from '../lib/auth'
import { roleLabels, Roles } from '../lib/roles'
import type { Role } from '../lib/roles'
import './UsersPage.css'

type CreateForm = {
  name: string
  email: string
  password: string
  role: Role
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Something went wrong'
}

export default function UsersPage() {
  const { user: me } = useAuth()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [form, setForm] = useState<CreateForm>({
    name: '',
    email: '',
    password: '',
    role: Roles.teacher,
  })
  const [creating, setCreating] = useState(false)

  const [editingId, setEditingId] = useState<number | null>(null)
  const [editName, setEditName] = useState('')
  const [editRole, setEditRole] = useState<Role>(Roles.teacher)

  const [resettingId, setResettingId] = useState<number | null>(null)
  const [newPassword, setNewPassword] = useState('')

  const isSuperAdmin = me?.role === Roles.superAdmin

  const assignableRoles = useMemo<Role[]>(
    () => (isSuperAdmin ? [Roles.admin, Roles.teacher, Roles.student] : [Roles.teacher, Roles.student]),
    [isSuperAdmin],
  )

  useEffect(() => {
    let active = true
    apiRequest<User[]>('/api/admin/users')
      .then((data) => {
        if (active) setUsers(data)
      })
      .catch((err: unknown) => {
        if (active) setError(errorMessage(err))
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => {
      active = false
    }
  }, [])

  if (me === null) return null

  const canEdit = (target: User): boolean => {
    if (target.role === Roles.superAdmin) {
      return me.role === Roles.superAdmin && me.id === target.id
    }
    if (target.role === Roles.admin) {
      return me.role === Roles.superAdmin
    }
    return true
  }

  const canDelete = (target: User): boolean => {
    if (target.role === Roles.superAdmin || target.id === me.id) return false
    if (target.role === Roles.admin) return me.role === Roles.superAdmin
    return true
  }

  const handleCreate = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setCreating(true)
    setError(null)
    try {
      const created = await apiRequest<User>('/api/admin/users', {
        method: 'POST',
        body: JSON.stringify(form),
      })
      setUsers((prev) => [...prev, created])
      setForm({ name: '', email: '', password: '', role: form.role })
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setCreating(false)
    }
  }

  const startEdit = (target: User) => {
    setEditingId(target.id)
    setEditName(target.name)
    setEditRole(target.role)
    setResettingId(null)
  }

  const cancelEdit = () => {
    setEditingId(null)
    setEditName('')
  }

  const saveEdit = async (target: User) => {
    setError(null)
    const payload: { name: string; role?: Role } = { name: editName.trim() }
    if (target.role !== Roles.superAdmin && target.id !== me.id && editRole !== target.role) {
      payload.role = editRole
    }
    try {
      const updated = await apiRequest<User>(`/api/admin/users/${target.id}`, {
        method: 'PATCH',
        body: JSON.stringify(payload),
      })
      setUsers((prev) => prev.map((u) => (u.id === updated.id ? updated : u)))
      cancelEdit()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  const startReset = (target: User) => {
    setResettingId(target.id)
    setNewPassword('')
    setEditingId(null)
  }

  const saveReset = async (event: FormEvent<HTMLFormElement>, target: User) => {
    event.preventDefault()
    setError(null)
    try {
      await apiRequest<null>(`/api/admin/users/${target.id}/password`, {
        method: 'POST',
        body: JSON.stringify({ password: newPassword }),
      })
      setResettingId(null)
      setNewPassword('')
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  const handleDelete = async (target: User) => {
    if (!window.confirm(`Delete ${target.name} (${target.email})?`)) return
    setError(null)
    try {
      await apiRequest<null>(`/api/admin/users/${target.id}`, { method: 'DELETE' })
      setUsers((prev) => prev.filter((u) => u.id !== target.id))
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <section className="users">
      <div className="users__head">
        <h1>Users</h1>
        <p>Create and manage accounts.</p>
      </div>

      <form className="users__create" onSubmit={handleCreate}>
        <h2>New user</h2>
        <div className="users__grid">
          <label className="users__field">
            <span>Name</span>
            <input
              value={form.name}
              onChange={(event) => setForm({ ...form, name: event.target.value })}
              required
            />
          </label>
          <label className="users__field">
            <span>Email</span>
            <input
              type="email"
              value={form.email}
              onChange={(event) => setForm({ ...form, email: event.target.value })}
              required
            />
          </label>
          <label className="users__field">
            <span>Password</span>
            <input
              type="password"
              autoComplete="new-password"
              value={form.password}
              onChange={(event) => setForm({ ...form, password: event.target.value })}
              minLength={8}
              required
            />
          </label>
          <label className="users__field">
            <span>Role</span>
            <select
              value={form.role}
              onChange={(event) => setForm({ ...form, role: event.target.value as Role })}
            >
              {assignableRoles.map((role) => (
                <option key={role} value={role}>
                  {roleLabels[role]}
                </option>
              ))}
            </select>
          </label>
        </div>
        <button type="submit" className="users__primary" disabled={creating}>
          {creating ? 'Creating…' : 'Create user'}
        </button>
      </form>

      {error && (
        <p className="users__error" role="alert">
          {error}
        </p>
      )}

      {loading ? (
        <p className="users__muted">Loading users…</p>
      ) : (
        <div className="users__table-wrap">
          <table className="users__table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Email</th>
              <th>Role</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {users.map((target) => {
              const editing = editingId === target.id
              const resetting = resettingId === target.id
              return (
                <tr key={target.id}>
                  <td>
                    {editing ? (
                      <input
                        className="users__inline"
                        value={editName}
                        onChange={(event) => setEditName(event.target.value)}
                      />
                    ) : (
                      target.name
                    )}
                  </td>
                  <td>{target.email}</td>
                  <td>
                    {editing ? (
                      <select
                        className="users__inline"
                        value={editRole}
                        disabled={target.role === Roles.superAdmin || target.id === me.id}
                        onChange={(event) => setEditRole(event.target.value as Role)}
                      >
                        {(target.role === Roles.superAdmin
                          ? ([Roles.superAdmin] as Role[])
                          : assignableRoles
                        ).map((role) => (
                          <option key={role} value={role}>
                            {roleLabels[role]}
                          </option>
                        ))}
                      </select>
                    ) : (
                      roleLabels[target.role]
                    )}
                  </td>
                  <td className="users__actions">
                    {editing ? (
                      <>
                        <button
                          type="button"
                          className="users__primary"
                          onClick={() => void saveEdit(target)}
                        >
                          Save
                        </button>
                        <button type="button" className="users__ghost" onClick={cancelEdit}>
                          Cancel
                        </button>
                      </>
                    ) : resetting ? (
                      <form className="users__reset" onSubmit={(event) => void saveReset(event, target)}>
                        <input
                          type="password"
                          className="users__inline"
                          placeholder="New password"
                          autoComplete="new-password"
                          value={newPassword}
                          minLength={8}
                          onChange={(event) => setNewPassword(event.target.value)}
                          required
                        />
                        <button type="submit" className="users__primary">
                          Set
                        </button>
                        <button
                          type="button"
                          className="users__ghost"
                          onClick={() => setResettingId(null)}
                        >
                          Cancel
                        </button>
                      </form>
                    ) : (
                      <>
                        {canEdit(target) && (
                          <>
                            <button type="button" className="users__ghost" onClick={() => startEdit(target)}>
                              Edit
                            </button>
                            <button
                              type="button"
                              className="users__ghost"
                              onClick={() => startReset(target)}
                            >
                              Reset password
                            </button>
                          </>
                        )}
                        {canDelete(target) && (
                          <button
                            type="button"
                            className="users__danger"
                            onClick={() => void handleDelete(target)}
                          >
                            Delete
                          </button>
                        )}
                      </>
                    )}
                  </td>
                </tr>
              )
            })}
          </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
