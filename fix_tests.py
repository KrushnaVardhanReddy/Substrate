import re

with open("api/internal/handlers/insurance_test.go", "r") as f:
    content = f.read()

# Add GetOrgIDByNameFunc to mockStore in TestInsuranceGetPolicyHandler
old1 = """	mockStore := &db.MockStore{
		GetInsurancePolicyFunc: func(ctx context.Context, id uuid.UUID) (*db.InsurancePolicy, error) {"""
new1 = """	mockStore := &db.MockStore{
		GetOrgIDByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			return orgID, nil
		},
		GetInsurancePolicyFunc: func(ctx context.Context, id uuid.UUID) (*db.InsurancePolicy, error) {"""
content = content.replace(old1, new1)

# Add GetOrgIDByNameFunc to mockStore in TestInsuranceGetClaimsHandler
old2 = """	mockStore := &db.MockStore{
		GetInsuranceClaimsFunc: func(ctx context.Context, id uuid.UUID) ([]db.InsuranceClaim, error) {"""
new2 = """	mockStore := &db.MockStore{
		GetOrgIDByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			return orgID, nil
		},
		GetInsuranceClaimsFunc: func(ctx context.Context, id uuid.UUID) ([]db.InsuranceClaim, error) {"""
content = content.replace(old2, new2)

with open("api/internal/handlers/insurance_test.go", "w") as f:
    f.write(content)
