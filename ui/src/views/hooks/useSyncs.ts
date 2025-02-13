import { useQuery } from '@apollo/client'
import { useRouter } from 'next/router'
import { GetRepoSyncsQuery } from 'src/api-logic/graphql/generated/schema'
import { GET_REPO_SYNCS } from 'src/api-logic/graphql/queries/get-repo-syncs.query'
import { mapToSyncsData } from 'src/api-logic/mappers/repo-syncs'
import { MERGESTAT_TITLE } from 'src/utils/constants'

const useSyncs = () => {
  const router = useRouter()
  const { repository } = router.query

  const { loading, data } = useQuery<GetRepoSyncsQuery>(GET_REPO_SYNCS, {
    variables: { id: repository },
    fetchPolicy: 'no-cache',
    pollInterval: 5000,
  })

  const repo = mapToSyncsData(data)
  const title = `${repo?.name} - Repos ${MERGESTAT_TITLE}`

  return { loading, repo, title }
}

export default useSyncs
