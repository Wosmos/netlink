import { useState } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  Alert,
  Switch,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Colors } from '@/constants/Colors';
import { useAuthStore } from '@/store/authStore';
import { APP_CONFIG } from '@/constants/Config';

export default function SettingsScreen() {
  const { user, logout } = useAuthStore();
  const [notifications, setNotifications] = useState(true);
  const [soundEnabled, setSoundEnabled] = useState(true);
  const [vibrationEnabled, setVibrationEnabled] = useState(true);

  const handleLogout = () => {
    Alert.alert(
      'Logout',
      'Are you sure you want to logout?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Logout',
          style: 'destructive',
          onPress: logout,
        },
      ]
    );
  };

  const handleDeleteAccount = () => {
    Alert.alert(
      'Delete Account',
      'This action cannot be undone. All your data will be permanently deleted.',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete Account',
          style: 'destructive',
          onPress: () => {
            Alert.alert('Feature Coming Soon', 'Account deletion will be available in a future update.');
          },
        },
      ]
    );
  };

  const renderSection = (title: string, children: React.ReactNode) => (
    <View style={styles.section}>
      <Text style={styles.sectionTitle}>{title}</Text>
      {children}
    </View>
  );

  const renderSettingItem = (
    icon: string,
    title: string,
    subtitle?: string,
    onPress?: () => void,
    rightComponent?: React.ReactNode,
    danger?: boolean
  ) => (
    <TouchableOpacity
      style={styles.settingItem}
      onPress={onPress}
      disabled={!onPress}
    >
      <View style={styles.settingLeft}>
        <Ionicons 
          name={icon as any} 
          size={20} 
          color={danger ? Colors.error : Colors.primary} 
        />
        <View style={styles.settingText}>
          <Text style={[styles.settingTitle, danger && styles.dangerText]}>
            {title}
          </Text>
          {subtitle && (
            <Text style={styles.settingSubtitle}>{subtitle}</Text>
          )}
        </View>
      </View>
      {rightComponent || (onPress && (
        <Ionicons name="chevron-forward" size={16} color={Colors.textMuted} />
      ))}
    </TouchableOpacity>
  );

  return (
    <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>SYSTEM CONFIG</Text>
      </View>

      {/* User Profile */}
      {renderSection('USER PROFILE', (
        <View style={styles.profileCard}>
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>
              {user?.name?.charAt(0) || user?.email?.charAt(0) || 'U'}
            </Text>
          </View>
          <View style={styles.profileInfo}>
            <Text style={styles.profileName}>
              {user?.name || 'Unknown User'}
            </Text>
            <Text style={styles.profileEmail}>{user?.email}</Text>
            {user?.is_verified ? (
              <View style={styles.verifiedBadge}>
                <Ionicons name="checkmark-circle" size={12} color={Colors.success} />
                <Text style={styles.verifiedText}>VERIFIED</Text>
              </View>
            ) : (
              <View style={styles.unverifiedBadge}>
                <Ionicons name="alert-circle" size={12} color={Colors.warning} />
                <Text style={styles.unverifiedText}>UNVERIFIED</Text>
              </View>
            )}
          </View>
        </View>
      ))}

      {/* Notifications */}
      {renderSection('NOTIFICATIONS', (
        <>
          {renderSettingItem(
            'notifications-outline',
            'Push Notifications',
            'Receive notifications for new messages',
            undefined,
            <Switch
              value={notifications}
              onValueChange={setNotifications}
              trackColor={{ false: Colors.textMuted, true: Colors.primary }}
              thumbColor={Colors.text}
            />
          )}
          {renderSettingItem(
            'volume-high-outline',
            'Sound',
            'Play sound for notifications',
            undefined,
            <Switch
              value={soundEnabled}
              onValueChange={setSoundEnabled}
              trackColor={{ false: Colors.textMuted, true: Colors.primary }}
              thumbColor={Colors.text}
            />
          )}
          {renderSettingItem(
            'phone-portrait-outline',
            'Vibration',
            'Vibrate for notifications',
            undefined,
            <Switch
              value={vibrationEnabled}
              onValueChange={setVibrationEnabled}
              trackColor={{ false: Colors.textMuted, true: Colors.primary }}
              thumbColor={Colors.text}
            />
          )}
        </>
      ))}

      {/* Privacy & Security */}
      {renderSection('PRIVACY & SECURITY', (
        <>
          {renderSettingItem(
            'lock-closed-outline',
            'Change Password',
            'Update your account password',
            () => Alert.alert('Feature Coming Soon', 'Password change will be available in a future update.')
          )}
          {renderSettingItem(
            'shield-outline',
            'Two-Factor Authentication',
            'Add an extra layer of security',
            () => Alert.alert('Feature Coming Soon', '2FA will be available in a future update.')
          )}
          {renderSettingItem(
            'eye-outline',
            'Privacy Settings',
            'Control who can see your information',
            () => Alert.alert('Feature Coming Soon', 'Privacy settings will be available in a future update.')
          )}
        </>
      ))}

      {/* Data & Storage */}
      {renderSection('DATA & STORAGE', (
        <>
          {renderSettingItem(
            'download-outline',
            'Export Data',
            'Download your messages and files',
            () => Alert.alert('Feature Coming Soon', 'Data export will be available in a future update.')
          )}
          {renderSettingItem(
            'trash-outline',
            'Clear Cache',
            'Free up storage space',
            () => {
              Alert.alert(
                'Clear Cache',
                'This will clear temporary files and cached data.',
                [
                  { text: 'Cancel', style: 'cancel' },
                  { text: 'Clear', onPress: () => Alert.alert('Success', 'Cache cleared successfully') },
                ]
              );
            }
          )}
        </>
      ))}

      {/* Support */}
      {renderSection('SUPPORT', (
        <>
          {renderSettingItem(
            'help-circle-outline',
            'Help & FAQ',
            'Get help and find answers',
            () => Alert.alert('Feature Coming Soon', 'Help section will be available in a future update.')
          )}
          {renderSettingItem(
            'mail-outline',
            'Contact Support',
            'Get in touch with our team',
            () => Alert.alert('Feature Coming Soon', 'Contact support will be available in a future update.')
          )}
          {renderSettingItem(
            'star-outline',
            'Rate App',
            'Rate NetLink on the App Store',
            () => Alert.alert('Feature Coming Soon', 'App rating will be available in a future update.')
          )}
        </>
      ))}

      {/* About */}
      {renderSection('ABOUT', (
        <>
          {renderSettingItem(
            'information-circle-outline',
            'Version',
            `${APP_CONFIG.VERSION} (${APP_CONFIG.BUILD_NUMBER})`
          )}
          {renderSettingItem(
            'document-text-outline',
            'Terms of Service',
            'Read our terms and conditions',
            () => Alert.alert('Feature Coming Soon', 'Terms of service will be available in a future update.')
          )}
          {renderSettingItem(
            'shield-checkmark-outline',
            'Privacy Policy',
            'Learn how we protect your data',
            () => Alert.alert('Feature Coming Soon', 'Privacy policy will be available in a future update.')
          )}
        </>
      ))}

      {/* Danger Zone */}
      {renderSection('DANGER ZONE', (
        <>
          {renderSettingItem(
            'log-out-outline',
            'Logout',
            'Sign out of your account',
            handleLogout,
            undefined,
            true
          )}
          {renderSettingItem(
            'trash-outline',
            'Delete Account',
            'Permanently delete your account',
            handleDeleteAccount,
            undefined,
            true
          )}
        </>
      ))}

      <View style={styles.footer}>
        <Text style={styles.footerText}>
          NetLink v{APP_CONFIG.VERSION}
        </Text>
        <Text style={styles.footerSubtext}>
          Secure Communication Protocol
        </Text>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.background,
  },
  header: {
    paddingHorizontal: 20,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: Colors.border,
  },
  headerTitle: {
    fontSize: 16,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.primary,
    letterSpacing: 2,
  },
  section: {
    marginTop: 24,
  },
  sectionTitle: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.textMuted,
    letterSpacing: 2,
    paddingHorizontal: 20,
    marginBottom: 12,
  },
  profileCard: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: Colors.surfaceVariant,
    marginHorizontal: 20,
    padding: 16,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: Colors.border,
  },
  avatar: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: Colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 16,
  },
  avatarText: {
    fontSize: 18,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.background,
  },
  profileInfo: {
    flex: 1,
  },
  profileName: {
    fontSize: 16,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.text,
    marginBottom: 4,
  },
  profileEmail: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
    marginBottom: 8,
  },
  verifiedBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: Colors.success + '20',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
    alignSelf: 'flex-start',
  },
  verifiedText: {
    fontSize: 10,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.success,
    marginLeft: 4,
    letterSpacing: 1,
  },
  unverifiedBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: Colors.warning + '20',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
    alignSelf: 'flex-start',
  },
  unverifiedText: {
    fontSize: 10,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.warning,
    marginLeft: 4,
    letterSpacing: 1,
  },
  settingItem: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 20,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: Colors.border,
  },
  settingLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
  },
  settingText: {
    marginLeft: 16,
    flex: 1,
  },
  settingTitle: {
    fontSize: 14,
    fontFamily: 'SpaceMono',
    color: Colors.text,
    marginBottom: 2,
  },
  settingSubtitle: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
  },
  dangerText: {
    color: Colors.error,
  },
  footer: {
    alignItems: 'center',
    paddingVertical: 32,
    paddingHorizontal: 20,
  },
  footerText: {
    fontSize: 12,
    fontFamily: 'SpaceMono',
    fontWeight: 'bold',
    color: Colors.primary,
    letterSpacing: 2,
    marginBottom: 4,
  },
  footerSubtext: {
    fontSize: 10,
    fontFamily: 'SpaceMono',
    color: Colors.textMuted,
    letterSpacing: 1,
  },
});