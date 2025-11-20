// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

output "account_id" {
  description = "AWS Account ID"
  value       = module.hello.account_id
}

output "arn" {
  description = "AWS Caller Identity ARN"
  value       = module.hello.arn
}

output "hello_message" {
  description = "Hello message"
  value       = module.hello.hello_message
}
