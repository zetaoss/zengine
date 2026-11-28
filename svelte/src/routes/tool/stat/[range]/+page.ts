import { redirect } from '@sveltejs/kit'

import type { PageLoad } from './$types'

// The ranges are 48h and 240d. Earlier ones (28d, 180d) and anything else go to the nearest of them.
export const load: PageLoad = ({ params }) => {
  if (params.range === '48h' || params.range === '240d') return
  throw redirect(302, params.range.endsWith('d') ? '/tool/stat/240d' : '/tool/stat/48h')
}
