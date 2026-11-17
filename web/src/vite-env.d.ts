/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** 测试构建开关：只有为 '1' 时登录页才预填测试凭据。 */
  readonly VITE_TEST_LOGIN?: string
  /** 测试账号，由 scripts/build-test.sh 注入。 */
  readonly VITE_TEST_USERNAME?: string
  /** 测试密码，由 scripts/build-test.sh 注入。 */
  readonly VITE_TEST_PASSWORD?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
