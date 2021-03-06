/**
 * @generated
 * Copyright 2004-present Facebook. All Rights Reserved.
 *
 **/

 /**
 * @flow
 */

/* eslint-disable */

'use strict';

/*::
import type { ReaderFragment } from 'relay-runtime';
type EquipmentBreadcrumbs_equipment$ref = any;
import type { FragmentReference } from "relay-runtime";
declare export opaque type ServiceEndpointsView_endpoints$ref: FragmentReference;
declare export opaque type ServiceEndpointsView_endpoints$fragmentType: ServiceEndpointsView_endpoints$ref;
export type ServiceEndpointsView_endpoints = $ReadOnlyArray<{|
  +id: string,
  +port: ?{|
    +parentEquipment: {|
      +name: string,
      +$fragmentRefs: EquipmentBreadcrumbs_equipment$ref,
    |},
    +definition: {|
      +id: string,
      +name: string,
    |},
  |},
  +equipment: {|
    +name: string,
    +$fragmentRefs: EquipmentBreadcrumbs_equipment$ref,
  |},
  +definition: {|
    +name: string,
    +role: ?string,
  |},
  +$refType: ServiceEndpointsView_endpoints$ref,
|}>;
export type ServiceEndpointsView_endpoints$data = ServiceEndpointsView_endpoints;
export type ServiceEndpointsView_endpoints$key = $ReadOnlyArray<{
  +$data?: ServiceEndpointsView_endpoints$data,
  +$fragmentRefs: ServiceEndpointsView_endpoints$ref,
  ...
}>;
*/


const node/*: ReaderFragment*/ = (function(){
var v0 = {
  "kind": "ScalarField",
  "alias": null,
  "name": "id",
  "args": null,
  "storageKey": null
},
v1 = {
  "kind": "ScalarField",
  "alias": null,
  "name": "name",
  "args": null,
  "storageKey": null
},
v2 = [
  (v1/*: any*/),
  {
    "kind": "FragmentSpread",
    "name": "EquipmentBreadcrumbs_equipment",
    "args": null
  }
];
return {
  "kind": "Fragment",
  "name": "ServiceEndpointsView_endpoints",
  "type": "ServiceEndpoint",
  "metadata": {
    "plural": true
  },
  "argumentDefinitions": [],
  "selections": [
    (v0/*: any*/),
    {
      "kind": "LinkedField",
      "alias": null,
      "name": "port",
      "storageKey": null,
      "args": null,
      "concreteType": "EquipmentPort",
      "plural": false,
      "selections": [
        {
          "kind": "LinkedField",
          "alias": null,
          "name": "parentEquipment",
          "storageKey": null,
          "args": null,
          "concreteType": "Equipment",
          "plural": false,
          "selections": (v2/*: any*/)
        },
        {
          "kind": "LinkedField",
          "alias": null,
          "name": "definition",
          "storageKey": null,
          "args": null,
          "concreteType": "EquipmentPortDefinition",
          "plural": false,
          "selections": [
            (v0/*: any*/),
            (v1/*: any*/)
          ]
        }
      ]
    },
    {
      "kind": "LinkedField",
      "alias": null,
      "name": "equipment",
      "storageKey": null,
      "args": null,
      "concreteType": "Equipment",
      "plural": false,
      "selections": (v2/*: any*/)
    },
    {
      "kind": "LinkedField",
      "alias": null,
      "name": "definition",
      "storageKey": null,
      "args": null,
      "concreteType": "ServiceEndpointDefinition",
      "plural": false,
      "selections": [
        (v1/*: any*/),
        {
          "kind": "ScalarField",
          "alias": null,
          "name": "role",
          "args": null,
          "storageKey": null
        }
      ]
    }
  ]
};
})();
// prettier-ignore
(node/*: any*/).hash = 'c9bdfa0f5793f7884b71103e23e3c420';
module.exports = node;
