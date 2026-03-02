import { useEffect, useRef, useState } from 'react';
import * as d3 from 'd3';
import type { NetworkTopology, TopologyNode, TopologyLink } from '../api/types';
import { useTheme } from '../contexts/ThemeContext';

interface NetworkTopologyProps {
  data: NetworkTopology | undefined;
  isLoading: boolean;
  error: unknown;
}

const NetworkTopology: React.FC<NetworkTopologyProps> = ({ data, isLoading, error }) => {
  const svgRef = useRef<SVGSVGElement>(null);
  const [selectedNode, setSelectedNode] = useState<TopologyNode | null>(null);
  const { theme } = useTheme();
  const isDark = theme === 'dark';

  useEffect(() => {
    if (!data || !svgRef.current) return;

    // Clear previous visualization
    d3.select(svgRef.current).selectAll('*').remove();

    const textColor = isDark ? '#e5e7eb' : '#334155';
    const strokeColor = isDark ? '#374151' : '#ffffff';

    const svg = d3.select(svgRef.current);
    const width = svgRef.current.clientWidth;
    const height = svgRef.current.clientHeight;

    // Create simulation
    const simulation = d3
      .forceSimulation(data.nodes as d3.SimulationNodeDatum[])
      .force('link', d3.forceLink(data.links as any)
        .id((d: any) => d.id)
        .distance(100))
      .force('charge', d3.forceManyBody().strength(-300))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(30));

    // Create container group
    const g = svg.append('g');

    // Add zoom behavior
    const zoom = d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.1, 4])
      .on('zoom', (event) => {
        g.attr('transform', event.transform);
      });
    svg.call(zoom as any);

    // Create links
    const link = g.append('g')
      .attr('class', 'links')
      .selectAll('line')
      .data(data.links)
      .enter()
      .append('line')
      .attr('stroke', (d: TopologyLink) => {
        switch (d.type) {
          case 'validates': return '#10b981'; // green
          case 'submits_to': return '#3b82f6'; // blue
          case 'votes_for': return '#f59e0b'; // amber
          default: return '#94a3b8';
        }
      })
      .attr('stroke-opacity', 0.6)
      .attr('stroke-width', 2);

    // Create nodes
    const node = g.append('g')
      .attr('class', 'nodes')
      .selectAll('g')
      .data(data.nodes)
      .enter()
      .append('g')
      .attr('class', 'node')
      .style('cursor', 'pointer')
      .call(d3.drag<SVGGElement, TopologyNode>()
        .on('start', dragstarted)
        .on('drag', dragged)
        .on('end', dragended) as any);

    // Add circles for nodes
    node.append('circle')
      .attr('r', (d: TopologyNode) => {
        switch (d.type) {
          case 'validator': return 20;
          case 'slot': return 15;
          case 'project': return 12;
          default: return 10;
        }
      })
      .attr('fill', (d: TopologyNode) => {
        switch (d.type) {
          case 'validator': return '#10b981'; // green
          case 'slot': return '#3b82f6'; // blue
          case 'project': return '#f59e0b'; // amber
          default: return '#94a3b8';
        }
      })
      .attr('stroke', strokeColor)
      .attr('stroke-width', 2)
      .on('mouseover', function() {
        d3.select(this).attr('stroke-width', 4);
      })
      .on('mouseout', function() {
        d3.select(this).attr('stroke-width', 2);
      })
      .on('click', (_event, d: TopologyNode) => {
        setSelectedNode(d);
      });

    // Add labels
    node.append('text')
      .text((d: TopologyNode) => d.label)
      .attr('x', (d: TopologyNode) => {
        switch (d.type) {
          case 'validator': return 25;
          case 'slot': return 20;
          case 'project': return 17;
          default: return 15;
        }
      })
      .attr('dy', '.35em')
      .attr('font-size', '12px')
      .attr('fill', textColor)
      .style('pointer-events', 'none');

    // Update positions on tick
    simulation.on('tick', () => {
      link
        .attr('x1', (d: any) => d.source.x)
        .attr('y1', (d: any) => d.source.y)
        .attr('x2', (d: any) => d.target.x)
        .attr('y2', (d: any) => d.target.y);

      node.attr('transform', (d: any) => `translate(${d.x},${d.y})`);
    });

    // Drag functions
    function dragstarted(event: d3.D3DragEvent<SVGGElement, TopologyNode, d3.SubjectPosition>, d: TopologyNode) {
      if (!event.active) simulation.alphaTarget(0.3).restart();
      d.fx = d.x ?? 0;
      d.fy = d.y ?? 0;
    }

    function dragged(event: d3.D3DragEvent<SVGGElement, TopologyNode, d3.SubjectPosition>, d: TopologyNode) {
      d.fx = event.x;
      d.fy = event.y;
    }

    function dragended(event: d3.D3DragEvent<SVGGElement, TopologyNode, d3.SubjectPosition>, d: TopologyNode) {
      if (!event.active) simulation.alphaTarget(0);
      d.fx = null;
      d.fy = null;
    }

    return () => {
      simulation.stop();
    };
  }, [data, isDark]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="text-red-500">Error loading network topology</div>
      </div>
    );
  }

  if (!data || data.nodes.length === 0) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="text-gray-500 dark:text-gray-400">No network data available</div>
      </div>
    );
  }

  return (
    <div className="relative w-full h-full">
      <svg
        ref={svgRef}
        className="w-full h-full border border-gray-200 dark:border-gray-600 rounded-lg"
        style={{ height: '600px', backgroundColor: isDark ? '#1f2937' : '#ffffff' }}
      />
      {selectedNode && (
        <div className="absolute top-4 right-4 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg p-4 shadow-lg max-w-xs">
          <div className="flex justify-between items-start mb-2">
            <h3 className="font-semibold text-lg text-gray-900 dark:text-white">{selectedNode.label}</h3>
            <button
              onClick={() => setSelectedNode(null)}
              className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            >
              ✕
            </button>
          </div>
          <div className="text-sm text-gray-600 dark:text-gray-300">
            <p><span className="font-medium">Type:</span> {selectedNode.type}</p>
            <p><span className="font-medium">ID:</span> {selectedNode.id}</p>
            {selectedNode.properties && Object.entries(selectedNode.properties).map(([key, value]) => (
              <p key={key}>
                <span className="font-medium">{key}:</span> {String(value)}
              </p>
            ))}
          </div>
        </div>
      )}
      <div className="absolute bottom-4 left-4 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg p-3 shadow-sm">
        <h4 className="font-medium text-sm mb-2 text-gray-900 dark:text-white">Legend</h4>
        <div className="flex gap-4 text-xs text-gray-600 dark:text-gray-300">
          <div className="flex items-center gap-1">
            <span className="w-3 h-3 rounded-full bg-green-500"></span>
            <span>Validator</span>
          </div>
          <div className="flex items-center gap-1">
            <span className="w-3 h-3 rounded-full bg-blue-500"></span>
            <span>Slot</span>
          </div>
          <div className="flex items-center gap-1">
            <span className="w-3 h-3 rounded-full bg-amber-500"></span>
            <span>Project</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default NetworkTopology;
