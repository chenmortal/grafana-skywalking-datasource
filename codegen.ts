import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: './graphql/schema.graphql',
  documents: ['./graphql/*.graphql'],
  generates: {
    // 只生成 schema 类型（包含 Duration, Step 等）
    // './src/generated/schema.ts': {
    //   plugins: ['typescript'],
    //   config: {
    //     scalars: { Long: 'number' },
    //     enumsAsTypes: false,
    //     skipTypename: true,
    //   },
    // },
    // 生成查询响应类型
    './src/type/operations.ts': {
      plugins: ['typescript-operations'],
      config: {
        scalars: { Long: 'number' },
        skipTypename: true,
      },
    },
  },
};

export default config;