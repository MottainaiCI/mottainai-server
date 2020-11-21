import { useContext } from "preact/hooks"
import ThemeContext from "@/contexts/theme"
import themes from "@/themes"

const DashboardCard = ({ title, children, loading, error }) => {
  let { theme } = useContext(ThemeContext)
  return (
    <div
      className={`w-full mb-4
        ${themes[theme].dashboard.cardBg}
        ${themes[theme].dashboard.cardBorder}`}
    >
      <div className={`p-2 text-lg ${themes[theme].dashboard.cardTitleBorder}`}>
        {title}
      </div>
      <div className="py-4">
        {error ? (
          <div className="text-center">There was an error loading data</div>
        ) : loading ? (
          <div className="text-center">Loading ...</div>
        ) : (
          children
        )}
      </div>
    </div>
  )
}

export default DashboardCard
