import fs from 'fs'
import React from 'react'

import { MainPage, PolicyNav, Markdown } from '../common'

export default (
  <MainPage title="Euphoria: Values" nav={<PolicyNav selected="values" />}>
    <Markdown className="text-page policy" content={fs.readFileSync(__dirname + '/values.md', 'utf8')} />
  </MainPage>
)
