import AppShell from './components/AppShell'
import LoginPage from './pages/LoginPage'
import { AuthProvider } from './lib/AuthProvider'
import { useAuth } from './lib/auth'

function Root() {
  const { user, loading } = useAuth()

  if (loading) {
    return <div className="app-loading">Loading…</div>
  }

  return user === null ? <LoginPage /> : <AppShell />
}

function App() {
  return (
    <AuthProvider>
      <Root />
    </AuthProvider>
  )
}

export default App
