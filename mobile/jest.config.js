/** @type {import('jest').Config} */
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
  },
  testMatch: [
    '**/__tests__/**/*.(test|spec).(ts|tsx|js)',
    '**/*.(test|spec).(ts|tsx|js)',
  ],
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/types.ts',
    '!src/**/index.ts',
    '!src/navigation/**',
    '!src/screens/**',
    '!src/components/**',
    '!src/assets/**',
    '!src/theme/**',
    '!src/hooks/**',
    '!src/services/backup/CloudBackupService.ts',
  ],
  coverageDirectory: '../tmp/mobile-coverage',
  setupFiles: ['<rootDir>/jest.setup.js'],
};
