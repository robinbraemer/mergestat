import { ApolloError } from '@apollo/client'
import { Avatar, Badge, Toolbar } from '@mergestat/blocks'
import { CircleWarningFilledIcon, CircleWarningIcon } from '@mergestat/icons'

type QueryEditorErrorProps = {
  errors?: ApolloError
  children?: React.ReactNode
}

const QueryEditorError: React.FC<QueryEditorErrorProps> = ({ errors }: QueryEditorErrorProps) => {
  return (
    <>
      <div className='bg-white flex h-16 w-full border-b px-8' style={{ minHeight: '64px' }}>
        <Toolbar className='flex-1 w-auto space-x-4 h-full'>
          <Toolbar.Left className='space-x-4'>
            <Toolbar.Item>
              <h3 className='t-h3 whitespace-nowrap mb-0'>Query results</h3>
            </Toolbar.Item>
            <Toolbar.Item>
              <Badge
                label='Error'
                variant='default'
                className='bg-gray-50 rounded-full'
                startIcon={
                  <CircleWarningFilledIcon className='t-icon t-icon-danger ' />
                }
              />
            </Toolbar.Item>
          </Toolbar.Left>
        </Toolbar>
      </div>
      <div className='bg-white w-full overflow-hidden flex-1 flex flex-col items-center justify-center'>
        <Avatar
          variant='danger'
          icon={<CircleWarningIcon className='t-icon' />}
        />
        <div className='text-center mt-6'>
          <h4 className='t-h4'>Error executing query</h4>
          <p className='t-text-muted'>
            {errors?.message}
          </p>
        </div>
      </div>
    </>
  )
}

export default QueryEditorError
