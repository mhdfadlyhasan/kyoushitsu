export const Roles = {
  superAdmin: 'super_admin',
  admin: 'admin',
  teacher: 'teacher',
  student: 'student',
} as const

export type Role = (typeof Roles)[keyof typeof Roles]

export const roleLabels: Record<Role, string> = {
  [Roles.superAdmin]: 'Super admin',
  [Roles.admin]: 'Admin',
  [Roles.teacher]: 'Teacher',
  [Roles.student]: 'Student',
}
