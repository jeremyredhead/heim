import fs from 'fs'
import React from 'react'

import { MainPage, PolicyNav, Markdown } from '../common'

export default (
  <MainPage title="Euphoria: Room Host Policy" nav={<PolicyNav selected="hosts" />}>
    <Markdown className="text-page policy" content={fs.readFileSync(__dirname + '/hosts.md', 'utf8')} />
  </MainPage>
)
