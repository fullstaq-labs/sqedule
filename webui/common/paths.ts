export function pathToApprovalRuleset(ruleset: any): string {
  var result = `/approval-rulesets/${encodeURIComponent(ruleset.id)}`;
  // if (ruleset.latest_approved_version) {
  //   result += `/versions/${encodeURIComponent(ruleset.latest_approved_version.version_number)}`;
  // }
  return result;
}
