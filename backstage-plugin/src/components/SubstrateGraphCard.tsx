import React, { useEffect, useState } from 'react';
import ReactFlow, { Background, Controls, Node, Edge } from 'react-flow-renderer';
import { Card, CardContent, Typography, CircularProgress } from '@material-ui/core';

interface SubstrateGraphCardProps {
  entityId: string;
}

export const SubstrateGraphCard = ({ entityId }: SubstrateGraphCardProps) => {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [edges, setEdges] = useState<Edge[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchGraph = async () => {
      try {
        setLoading(true);
        const response = await fetch(`/api/v1/registry/graph/${entityId}`);
        if (!response.ok) {
          throw new Error('Failed to fetch dependency graph');
        }

        const rawEdges = await response.json();

        // Convert flat edge array to React Flow nodes and edges
        const uniqueNodes = new Set<string>();
        const flowNodes: Node[] = [];
        const flowEdges: Edge[] = [];

        // Extract unique nodes
        rawEdges.forEach((edge: any) => {
          if (edge.provider) uniqueNodes.add(edge.provider);
          if (edge.consumer) uniqueNodes.add(edge.consumer);
        });

        // Create nodes
        Array.from(uniqueNodes).forEach((id, index) => {
          flowNodes.push({
            id,
            data: { label: id },
            position: { x: (index % 3) * 200, y: Math.floor(index / 3) * 100 },
          });
        });

        // Create edges
        rawEdges.forEach((edge: any) => {
          if (edge.provider && edge.consumer) {
            flowEdges.push({
              id: `${edge.consumer}-${edge.provider}`,
              source: edge.consumer,
              target: edge.provider,
              animated: edge.status === 'active',
            });
          }
        });

        setNodes(flowNodes);
        setEdges(flowEdges);
      } catch (err: any) {
        setError(err.message || 'An error occurred');
      } finally {
        setLoading(false);
      }
    };

    if (entityId) {
      fetchGraph();
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

  return (
    <Card style={{ height: 400 }}>
      <CardContent style={{ height: '100%' }}>
        <Typography variant="h6" gutterBottom>
          Substrate Dependency Graph ({entityId})
        </Typography>
        <div style={{ height: 'calc(100% - 40px)' }} data-testid="react-flow-container">
          <ReactFlow nodes={nodes} edges={edges} fitView>
            <Background />
            <Controls />
          </ReactFlow>
        </div>
      </CardContent>
    </Card>
  );
};
