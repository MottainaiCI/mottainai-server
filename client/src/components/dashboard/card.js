import { useContext, useEffect, useState } from "preact/hooks"
import ThemeContext from "@/contexts/theme"
import themes from "@/themes"
import Spinner from "@/components/spinner"

const DashboardCard = ({ title, children, loading, error }) => {
  let { theme } = useContext(ThemeContext)
  let [isLoading, setIsLoading] = useState(true)
  useEffect(() => {
    let handle = setTimeout(() => {
      setIsLoading(loading)
    }, 500)
    return () => clearTimeout(handle)
  }, [loading])

  return (
    <div
      className={`w-full mb-4
        ${themes[theme].cardBg}
        ${themes[theme].cardBorder}`}
    >
      <div className={`p-2 text-lg ${themes[theme].dashboard.cardTitleBorder}`}>
        {title}
      </div>
      <div className="py-4">
        {error ? (
          <div className="text-center">There was an error loading data</div>
        ) : isLoading ? (
          <Spinner />
        ) : (
          children
        )}
      </div>
    </div>
  )
}

export default DashboardCard
