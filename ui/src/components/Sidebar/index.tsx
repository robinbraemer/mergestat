import { Sidebar } from '@mergestat/blocks'
import { CogIcon, DatabaseIcon, RepositoryIcon, TerminalIcon } from '@mergestat/icons'
import { useRouter } from 'next/router'
import React from 'react'

const SidebarView: React.FC = () => {
  const { pathname, push } = useRouter()

  const isSidebarActive = (path: string) => !!pathname.match(path)?.length

  return (
    <Sidebar>
      <Sidebar.Item
        label="Repos"
        active={isSidebarActive('repos')}
        onClick={() => push('/repos')}
        icon={<RepositoryIcon className='t-icon' />}
      />
      <Sidebar.Item
        label="Queries"
        active={isSidebarActive('queries')}
        onClick={() => push('/queries')}
        icon={<TerminalIcon className='t-icon' />}
      />
      <Sidebar.Item
        label="Connect"
        active={isSidebarActive('connect')}
        onClick={() => push('/connect')}
        icon={<DatabaseIcon className='t-icon' />}
      />
      <Sidebar.Divider />
      <Sidebar.Item
        label="Settings"
        active={isSidebarActive('settings')}
        onClick={() => push('/settings')}
        icon={<CogIcon className='t-icon' />}
      />
    </Sidebar>
  )
}

export default SidebarView
