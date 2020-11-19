const DashboardCard = ({title, children}) => (
  <div className="bg-white border border-accent w-full mb-4">
    <div className="p-2 text-lg border-b border-accent">{title}</div>
    <div className="py-4">
      {children}
    </div>
  </div>
)

export default DashboardCard
