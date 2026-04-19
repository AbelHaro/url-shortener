import { defineConfig, defineTransformer } from "orval";

const renameDtos = defineTransformer((input) => {
  const schemaPrefix = "DtosV1";

  const schemaEntries = Object.entries(input.components?.schemas ?? {});
  const schemaNameMap = Object.fromEntries(
    schemaEntries
      .filter(([name]) => name.startsWith(schemaPrefix))
      .map(([name]) => [name, name.slice(schemaPrefix.length)]),
  );

  const rewriteRefs = (value: unknown): unknown => {
    if (Array.isArray(value)) {
      return value.map(rewriteRefs);
    }

    if (!value || typeof value !== "object") {
      return value;
    }

    return Object.fromEntries(
      Object.entries(value as Record<string, unknown>).map(
        ([key, nestedValue]) => {
          if (key === "$ref" && typeof nestedValue === "string") {
            const matchedName = Object.keys(schemaNameMap).find((name) =>
              nestedValue.endsWith(`/components/schemas/${name}`),
            );

            if (matchedName) {
              return [
                key,
                nestedValue.replace(
                  matchedName,
                  schemaNameMap[matchedName as keyof typeof schemaNameMap],
                ),
              ];
            }
          }

          return [key, rewriteRefs(nestedValue)];
        },
      ),
    );
  };

  const schemas = Object.fromEntries(
    schemaEntries.map(([name, schema]) => [
      schemaNameMap[name] ?? name,
      schema,
    ]),
  );

  return {
    // @ts-expect-error - We know the structure of the input and output, but it's too complex for TypeScript to verify
    ...rewriteRefs(input),
    components: {
      ...input.components,
      schemas,
    },
  };
});

export default defineConfig({
  petstore: {
    input: {
      target: "../backend/docs/swagger.yaml",
      override: {
        transformer: renameDtos,
      },
    },
    output: {
      target: "./src/api/generated.ts",
      schemas: "./src/api/model",
      client: "react-query",
      override: {
        mutator: {
          path: "./src/api/fetcher.ts",
          name: "authFetch",
        },
      },
      baseUrl: {
        runtime: 'import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"',
      },
    },
  },
});
