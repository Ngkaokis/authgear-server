import { useMutation } from "@apollo/client";
import {
  CreateRoleMutationDocument,
  CreateRoleMutationMutation,
} from "./createRoleMutation.generated";
import { useCallback } from "react";

export function useCreateRoleMutation(): {
  createRole: (
    key: string,
    name: string,
    description?: string
  ) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutateFunction, { error, loading }] =
    useMutation<CreateRoleMutationMutation>(CreateRoleMutationDocument);

  const createRole = useCallback(
    async (key: string, name: string, description?: string) => {
      const result = await mutateFunction({
        variables: {
          key,
          name,
          description,
        },
      });
      const roleID = result.data?.createRole.role.id ?? null;
      return roleID;
    },
    [mutateFunction]
  );

  return { createRole, error, loading };
}
