import { redirect } from '@sveltejs/kit'

import type { PageLoad } from './$types'

export const load: PageLoad = () => {
  throw redirect(302, '/tool/stat/48h')
}
