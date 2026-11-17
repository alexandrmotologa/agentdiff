import assert from "node:assert";
import { getUserAge, User } from "./user.ts";

function testNormalUser() {
  const user: User = {
    id: "u1",
    name: "Alex",
    profile: { age: 30 }
  };
  assert.strictEqual(getUserAge(user), 30);
}

function testNullUser() {
  assert.strictEqual(getUserAge(null), 0);
  assert.strictEqual(getUserAge(undefined), 0);
}

function testMissingProfile() {
  const user: User = { id: "u2", name: "Bob" };
  assert.strictEqual(getUserAge(user), 0);
}

testNormalUser();
testNullUser();
testMissingProfile();
console.log("All TypeScript user age tests passed!");
