import React, { useEffect, useState } from 'react';
import { Card, CardContent, Typography, CircularProgress, Box } from '@material-ui/core';

interface SubstrateHealthCardProps {
  entityId: string;
}

interface HealthData {
  score: number;
  riskLevel: 'Low' | 'Medium' | 'High';
}

export const SubstrateHealthCard = ({ entityId }: SubstrateHealthCardProps) => {
  const [data, setData] = useState<HealthData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        setLoading(true);
        // Assuming there is an endpoint for health score, e.g., /api/v1/registry/health/{entityId}
        // Since it's not strictly specified, simulating a fetch to what would be the endpoint
        const response = await fetch(`/api/v1/registry/health/${entityId}`);
        if (!response.ok) {
          throw new Error('Failed to fetch health score');
        }
        const result = await response.json();
        setData(result);
      } catch (err: any) {
        setError(err.message || 'An error occurred');
      } finally {
        setLoading(false);
      }
    };

    if (entityId) {
      fetchHealth();
    }
  }, [entityId]);

  if (loading) {
    return (
      <Card>
        <CardContent>
          <CircularProgress />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent>
          <Typography color="error">{error}</Typography>
        </CardContent>
      </Card>
    );
  }

  const getColor = (level: string) => {
    switch (level) {
      case 'Low': return 'green';
      case 'Medium': return 'orange';
      case 'High': return 'red';
      default: return 'gray';
    }
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          API Breaking Change Risk
        </Typography>
        {data && (
          <Box display="flex" flexDirection="column" alignItems="center">
            <Typography variant="h2" style={{ color: getColor(data.riskLevel) }}>
              {data.score}
            </Typography>
            <Typography variant="subtitle1" style={{ color: getColor(data.riskLevel) }}>
              {data.riskLevel} Risk
            </Typography>
          </Box>
        )}
      </CardContent>
    </Card>
  );
};
