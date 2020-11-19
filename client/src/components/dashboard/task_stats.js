import DashboardCard from './card'

const Stat = ({label, num}) => (
  <div className="text-center mx-5">
    <div className="text-3xl font-medium">{num}</div>
    <div className="text-md">{label}</div>
  </div>
)

const TaskStatCard = () => (
  <DashboardCard title="Task Stats">
    <div className="flex justify-center">
      <Stat label="Total" num={0} />
      <Stat label="Running" num={0} />
      <Stat label="Waiting" num={0} />
      <Stat label="Succeeded" num={0} />
      <Stat label="Failed" num={0} />
    </div>
  </DashboardCard>
)

export default TaskStatCard
