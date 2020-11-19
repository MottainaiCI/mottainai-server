import { Router } from 'preact-router'
import { useLocalStorage, writeStorage } from '@rehooks/local-storage';

import Dashboard from '@/components/pages/dashboard'
import Plans from '@/components/pages/plans'
import Pipelines from '@/components/pages/pipelines'
import Tasks from '@/components/pages/tasks'
import Artefacts from '@/components/pages/artefacts'
import Sidebar from '@/components/sidebar'

import ThemeContext from '@/contexts/theme'
import themes from '@/themes'

const App = () => {
	const [theme, setTheme] = useLocalStorage('mottainai-theme', 'light')
	const themeValue = { theme, setTheme }
	return (<ThemeContext.Provider value={themeValue} >
		<div className={`flex h-screen ${themes[theme].bg} ${themes[theme].textColor}`}>
			<div className="w-60 flex-none flex flex-col">
				<Sidebar />
			</div>
			<div className="px-8 py-10 flex-1 overflow-auto">
				<Router>
					<Dashboard path="/" />
					<Plans path="/plans/" />
					<Pipelines path="/pipelines/" />
					<Tasks path="/tasks/" />
					<Artefacts path="/artefacts/" />
				</Router>
			</div>
		</div>
	</ThemeContext.Provider>)
}

export default App
