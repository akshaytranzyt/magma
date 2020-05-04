#!/usr/bin/env python3
# @generated AUTOGENERATED file. Do not Change!

from dataclasses import dataclass
from datetime import datetime
from gql.gql.datetime_utils import DATETIME_FIELD
from gql.gql.graphql_client import GraphqlClient
from gql.gql.client import OperationException
from gql.gql.reporter import FailedOperationException
from functools import partial
from numbers import Number
from typing import Any, Callable, List, Mapping, Optional
from time import perf_counter
from dataclasses_json import DataClassJsonMixin

from .property_fragment import PropertyFragment, QUERY as PropertyFragmentQuery
from .edit_equipment_port_input import EditEquipmentPortInput


QUERY: List[str] = PropertyFragmentQuery + ["""
mutation EditEquipmentPortMutation($input: EditEquipmentPortInput!) {
  editEquipmentPort(input: $input) {
    id
    properties {
      ...PropertyFragment
    }
    definition {
      id
      name
      portType {
        id
        name
      }
    }
    link {
      id
      services {
        id
      }
      properties {
        ...PropertyFragment
      }
    }
  }
}

"""]

@dataclass
class EditEquipmentPortMutation(DataClassJsonMixin):
    @dataclass
    class EditEquipmentPortMutationData(DataClassJsonMixin):
        @dataclass
        class EquipmentPort(DataClassJsonMixin):
            @dataclass
            class Property(PropertyFragment):
                pass

            @dataclass
            class EquipmentPortDefinition(DataClassJsonMixin):
                @dataclass
                class EquipmentPortType(DataClassJsonMixin):
                    id: str
                    name: str

                id: str
                name: str
                portType: Optional[EquipmentPortType]

            @dataclass
            class Link(DataClassJsonMixin):
                @dataclass
                class Service(DataClassJsonMixin):
                    id: str

                @dataclass
                class Property(PropertyFragment):
                    pass

                id: str
                services: List[Service]
                properties: List[Property]

            id: str
            properties: List[Property]
            definition: EquipmentPortDefinition
            link: Optional[Link]

        editEquipmentPort: EquipmentPort

    data: EditEquipmentPortMutationData

    @classmethod
    # fmt: off
    def execute(cls, client: GraphqlClient, input: EditEquipmentPortInput) -> EditEquipmentPortMutationData.EquipmentPort:
        # fmt: off
        variables = {"input": input}
        try:
            start_time = perf_counter()
            response_text = client.call(''.join(set(QUERY)), variables=variables)
            res = cls.from_json(response_text).data
            elapsed_time = perf_counter() - start_time
            client.reporter.log_successful_operation("EditEquipmentPortMutation", variables, elapsed_time)
            return res.editEquipmentPort
        except OperationException as e:
            raise FailedOperationException(
                client.reporter,
                e.err_msg,
                e.err_id,
                "EditEquipmentPortMutation",
                variables,
            )