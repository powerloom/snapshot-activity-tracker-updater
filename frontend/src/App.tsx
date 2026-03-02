import { ThemeProvider } from './contexts/ThemeContext'
import Dashboard from './pages/Dashboard'

function App() {
  return (
    <ThemeProvider>
      <Dashboard />
    </ThemeProvider>
  )
}

export default App
