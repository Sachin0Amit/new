import * as d3 from 'd3';

export interface ReasoningStep {
  id: string;
  parentIds: string[];
  status: 'pending' | 'in-progress' | 'complete' | 'anomaly';
  confidence: number;
  payload: string;
  timestamp: string;
  signature: string;
}

interface GraphNode extends d3.SimulationNodeDatum {
  id: string;
  status: string;
}

interface GraphLink extends d3.SimulationLinkDatum<GraphNode> {
  source: string;
  target: string;
  confidence: number;
}

export class ReasoningVisualizer {
  private svg: d3.Selection<SVGSVGElement, unknown, HTMLElement, any>;
  private simulation: d3.Simulation<GraphNode, GraphLink>;
  private nodes: GraphNode[] = [];
  private links: GraphLink[] = [];

  constructor(containerId: string) {
    const container = d3.select(`#${containerId}`);
    const width = (container.node() as HTMLElement).clientWidth;
    const height = 600;

    this.svg = container.append('svg')
      .attr('width', width)
      .attr('height', height)
      .style('background', '#0b0e11');

    const g = this.svg.append('g');

    // Zoom functionality
    this.svg.call(d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.1, 4])
      .on('zoom', (event: any) => {
        g.attr('transform', (event as any).transform);
      }));

    this.simulation = d3.forceSimulation<GraphNode>(this.nodes)
      .force('link', d3.forceLink<GraphNode, GraphLink>(this.links).id(d => d.id).distance(100))
      .force('charge', d3.forceManyBody().strength(-300))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .on('tick', () => this.ticked(g));
  }

  public addStep(step: ReasoningStep): void {
    const newNode: GraphNode = { id: step.id, status: step.status };
    this.nodes.push(newNode);

    step.parentIds.forEach(parentId => {
      this.links.push({
        source: parentId,
        target: step.id,
        confidence: step.confidence
      });
    });

    this.update();
  }

  private update(): void {
    this.simulation.nodes(this.nodes);
    const linkForce = this.simulation.force('link') as d3.ForceLink<GraphNode, GraphLink>;
    linkForce.links(this.links);
    this.simulation.alpha(1).restart();
  }

  private ticked(g: d3.Selection<SVGGElement, unknown, HTMLElement, any>): void {
    const nodeSelection = g.selectAll<SVGCircleElement, GraphNode>('circle').data(this.nodes, d => d.id);
    const linkSelection = g.selectAll<SVGLineElement, GraphLink>('line').data(this.links, d => `${d.source}-${d.target}`);

    // Update Links
    linkSelection.enter().append('line')
      .attr('stroke', '#4a4f55')
      .attr('stroke-width', d => d.confidence * 5)
      .merge(linkSelection)
      .attr('x1', d => (d.source as any).x)
      .attr('y1', d => (d.source as any).y)
      .attr('x2', d => (d.target as any).x)
      .attr('y2', d => (d.target as any).y);

    // Update Nodes
    const colors: Record<string, string> = {
      pending: '#808080',
      'in-progress': '#ffbf00',
      complete: '#4caf50',
      anomaly: '#f44336'
    };

    nodeSelection.enter().append('circle')
      .attr('r', 10)
      .attr('cursor', 'pointer')
      .on('click', (_event, d) => this.showDetails(d))
      .merge(nodeSelection)
      .attr('cx', d => d.x!)
      .attr('cy', d => d.y!)
      .attr('fill', d => colors[d.status] || '#fff');
  }

  private showDetails(node: GraphNode): void {
    console.log(`[VIS] Step Details:`, node.id);
  }
}
