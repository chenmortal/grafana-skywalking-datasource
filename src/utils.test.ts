// src/__tests__/utils.test.ts
import { descodeServiceID } from './utils';

describe('decodeServiceID', () => {
  it('should decode base64 service ID', () => {
    const testId = btoa('service-123') + '.1';
    expect(descodeServiceID(testId)).toBe('service-123');
  });

  it('should handle invalid base64', () => {
    const testId = 'invalid!!!.1';
    expect(descodeServiceID(testId)).toBe(testId);
  });
});