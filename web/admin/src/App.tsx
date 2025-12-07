import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Login from './pages/Login'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Deposits from './pages/Deposits'
import Withdrawals from './pages/Withdrawals'
import Accounts from './pages/Accounts'
import Assets from './pages/Assets'
import Risk from './pages/Risk'
import AddressPool from './pages/AddressPool'
import Settings from './pages/Settings'
import PrivateRoute from './components/PrivateRoute'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/"
          element={
            <PrivateRoute>
              <Layout />
            </PrivateRoute>
          }
        >
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="deposits" element={<Deposits />} />
          <Route path="withdrawals" element={<Withdrawals />} />
          <Route path="accounts" element={<Accounts />} />
          <Route path="assets" element={<Assets />} />
          <Route path="risk" element={<Risk />} />
          <Route path="address-pool" element={<AddressPool />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App

