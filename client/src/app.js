import { Router } from 'preact-router'

// Code-splitting is automated for `routes` directory
import Dashboard from '@/components/pages/dashboard'
import Plans from '@/components/pages/plans'
import Pipelines from '@/components/pages/pipelines'
import Tasks from '@/components/pages/tasks'
import Artefacts from '@/components/pages/artefacts'

import Sidebar from '@/components/sidebar'

const App = () => (
	<div className="bg-secondary theme-dark min-h-full flex flex-row w-full">
		<Sidebar />
		<div className="px-8 py-10 w-full">
			<Router>
				<Dashboard path="/" />
				<Plans path="/plans/" />
				<Pipelines path="/pipelines/" />
				<Tasks path="/tasks/" />
				<Artefacts path="/artefacts/" />
			</Router>
		</div>
	</div>
)

export default App
