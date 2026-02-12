// Auth Navigator — Phone → OTP → Profile Setup
import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import type { AuthStackParamList } from './types';
import { PhoneInputScreen } from '@/screens/auth/PhoneInputScreen';
import { OTPVerifyScreen } from '@/screens/auth/OTPVerifyScreen';
import { ReverseOTPWaitScreen } from '@/screens/auth/ReverseOTPWaitScreen';
import { ProfileSetupScreen } from '@/screens/auth/ProfileSetupScreen';
import { colors } from '@/theme';

const Stack = createNativeStackNavigator<AuthStackParamList>();

export function AuthNavigator() {
  return (
    <Stack.Navigator
      screenOptions={{
        headerShown: false,
        contentStyle: { backgroundColor: colors.background },
        animation: 'slide_from_right',
      }}
    >
      <Stack.Screen name="PhoneInput" component={PhoneInputScreen} />
      <Stack.Screen name="OTPVerify" component={OTPVerifyScreen} />
      <Stack.Screen name="ReverseOTPWait" component={ReverseOTPWaitScreen} />
      <Stack.Screen name="ProfileSetup" component={ProfileSetupScreen} />
    </Stack.Navigator>
  );
}
