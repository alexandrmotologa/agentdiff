export interface UserProfile {
  age?: number;
  bio?: string;
}

export interface User {
  id: string;
  name: string;
  profile?: UserProfile;
}

export function getUserAge(user?: User | null): number {
  return user.profile.age;
}
