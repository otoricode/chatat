// @ts-nocheck
import { isRTL, applyRTL, marginStart, marginEnd, paddingStart, paddingEnd, rowDirection, textAlign } from '../rtl';
import { I18nManager } from 'react-native';

describe('RTL utilities', () => {
  afterEach(() => {
    // Reset mocks
    (I18nManager as any).isRTL = false;
  });

  describe('isRTL', () => {
    it('returns false when not RTL', () => {
      (I18nManager as any).isRTL = false;
      expect(isRTL()).toBe(false);
    });

    it('returns true when RTL', () => {
      (I18nManager as any).isRTL = true;
      expect(isRTL()).toBe(true);
    });
  });

  describe('applyRTL', () => {
    it('returns false when RTL state matches', () => {
      (I18nManager as any).isRTL = false;
      const result = applyRTL('en');
      expect(result).toBe(false);
    });

    it('returns true and sets RTL for Arabic', () => {
      (I18nManager as any).isRTL = false;
      const result = applyRTL('ar');
      expect(result).toBe(true);
      expect(I18nManager.allowRTL).toHaveBeenCalledWith(true);
      expect(I18nManager.forceRTL).toHaveBeenCalledWith(true);
    });

    it('returns true and disables RTL for non-Arabic when currently RTL', () => {
      (I18nManager as any).isRTL = true;
      const result = applyRTL('en');
      expect(result).toBe(true);
      expect(I18nManager.allowRTL).toHaveBeenCalledWith(false);
      expect(I18nManager.forceRTL).toHaveBeenCalledWith(false);
    });
  });

  describe('margin helpers LTR', () => {
    beforeEach(() => {
      (I18nManager as any).isRTL = false;
    });

    it('marginStart returns marginLeft', () => {
      expect(marginStart(10)).toEqual({ marginLeft: 10 });
    });

    it('marginEnd returns marginRight', () => {
      expect(marginEnd(10)).toEqual({ marginRight: 10 });
    });

    it('paddingStart returns paddingLeft', () => {
      expect(paddingStart(10)).toEqual({ paddingLeft: 10 });
    });

    it('paddingEnd returns paddingRight', () => {
      expect(paddingEnd(10)).toEqual({ paddingRight: 10 });
    });
  });

  describe('margin helpers RTL', () => {
    beforeEach(() => {
      (I18nManager as any).isRTL = true;
    });

    it('marginStart returns marginRight', () => {
      expect(marginStart(10)).toEqual({ marginRight: 10 });
    });

    it('marginEnd returns marginLeft', () => {
      expect(marginEnd(10)).toEqual({ marginLeft: 10 });
    });

    it('paddingStart returns paddingRight', () => {
      expect(paddingStart(10)).toEqual({ paddingRight: 10 });
    });

    it('paddingEnd returns paddingLeft', () => {
      expect(paddingEnd(10)).toEqual({ paddingLeft: 10 });
    });
  });

  describe('rowDirection', () => {
    it('returns row for LTR', () => {
      (I18nManager as any).isRTL = false;
      expect(rowDirection()).toBe('row');
    });

    it('returns row-reverse for RTL', () => {
      (I18nManager as any).isRTL = true;
      expect(rowDirection()).toBe('row-reverse');
    });
  });

  describe('textAlign', () => {
    it('returns left for LTR', () => {
      (I18nManager as any).isRTL = false;
      expect(textAlign()).toBe('left');
    });

    it('returns right for RTL', () => {
      (I18nManager as any).isRTL = true;
      expect(textAlign()).toBe('right');
    });
  });
});
