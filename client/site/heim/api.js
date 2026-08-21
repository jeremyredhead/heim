import fs from 'fs'
import React from 'react'

import { MainPage, HeimNav, Markdown } from '../common'

export default (
  <MainPage title="Euphoria: API" nav={<HeimNav selected="api" />} sidebar>
    <Markdown className="text-page api" content={fs.readFileSync(__dirname + '/../../../doc/api.md', 'utf8')} />
  </MainPage>
)
