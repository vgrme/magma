/**
 * Copyright 2004-present Facebook. All Rights Reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 *
 * @flow
 * @format
 */

import type {ComponentType} from 'react';

type Section = {
  path: string,
  label: string,
  icon: any,
  component: ComponentType<any>,
};

export type SectionsConfigs = [?string, Section[]];
