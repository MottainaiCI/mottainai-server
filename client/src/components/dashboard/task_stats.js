import DashboardCard from "./card"
import { useEffect, useState } from "preact/hooks"
import axios from "redaxios"

const Stat = ({ label, num }) => (
  <div className="text-center mx-5">
    <div className="text-3xl font-medium">{num}</div>
    <div className="text-md">{label}</div>
  </div>
)

const TaskStatCard = () => {
  const [stats, setStats] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    axios("/api/client/dashboard/stats").then(
      (result) => {
        setStats(result)
        setLoading(false)
      },
      (error) => {
        setError(error)
      }
    )
  }, [])

  return (
    <DashboardCard title="Task Stats" loading={loading} error={error}>
      <div className="flex justify-center">
        <Stat label="Total" num={stats.total} />
        <Stat label="Running" num={stats.running} />
        <Stat label="Waiting" num={stats.waiting} />
        <Stat label="Succeeded" num={stats.error} />
        <Stat label="Failed" num={stats.failed} />
      </div>
    </DashboardCard>
  )
}

export default TaskStatCard
