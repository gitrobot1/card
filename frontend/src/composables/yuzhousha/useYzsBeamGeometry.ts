export interface BeamPoint {
  x1: number
  y1: number
  x2: number
  y2: number
}

export interface BeamPreviewLine extends BeamPoint {
  id: string
  label: string
  className: string
  midX: number
  midY: number
}

function seatAnchor(seat: number, root?: ParentNode | null) {
  const scope = root ?? document
  return scope.querySelector<HTMLElement>(`.ddz__seat-anchor[data-seat="${seat}"]`)
}

/** 射线与矩形边界的最近交点（从外侧射入），略向内缩避免压进头像 */
export function rayRectEdgePoint(
  ox: number,
  oy: number,
  tx: number,
  ty: number,
  rect: DOMRect,
  inset = 2,
): { x: number; y: number } {
  const dx = tx - ox
  const dy = ty - oy
  const len = Math.hypot(dx, dy) || 1
  const ux = dx / len
  const uy = dy / len

  let tMin = Infinity

  if (Math.abs(ux) > 1e-6) {
    for (const edge of [rect.left, rect.right]) {
      const t = (edge - ox) / ux
      if (t <= 0) continue
      const y = oy + uy * t
      if (y >= rect.top && y <= rect.bottom && t < tMin) tMin = t
    }
  }
  if (Math.abs(uy) > 1e-6) {
    for (const edge of [rect.top, rect.bottom]) {
      const t = (edge - oy) / uy
      if (t <= 0) continue
      const x = ox + ux * t
      if (x >= rect.left && x <= rect.right && t < tMin) tMin = t
    }
  }

  if (!Number.isFinite(tMin)) {
    return { x: tx, y: ty }
  }

  const edgeX = ox + ux * tMin
  const edgeY = oy + uy * tMin
  return {
    x: edgeX - ux * inset,
    y: edgeY - uy * inset,
  }
}

export function beamEndpointsForSeats(
  fromSeat: number,
  toSeat: number,
  humanSeat: number,
  handArea: HTMLElement | null,
  root?: ParentNode | null,
): BeamPoint | null {
  const fromEl = seatAnchor(fromSeat, root)
  const toEl = seatAnchor(toSeat, root)
  if (!fromEl || !toEl) return null

  const fromAnchor = fromSeat === humanSeat && handArea ? handArea : fromEl
  const fromRect = fromAnchor.getBoundingClientRect()
  const toRect = toEl.getBoundingClientRect()

  const x1 = fromRect.left + fromRect.width / 2
  const y1 = fromRect.top + fromRect.height * 0.38
  const targetCenterX = toRect.left + toRect.width / 2
  const targetCenterY = toRect.top + toRect.height / 2
  const edge = rayRectEdgePoint(x1, y1, targetCenterX, targetCenterY, toRect)

  return {
    x1,
    y1,
    x2: edge.x,
    y2: edge.y,
  }
}

export function offsetBeamPoint(base: BeamPoint, index: number, total: number, gap = 7): BeamPoint {
  const dx = base.x2 - base.x1
  const dy = base.y2 - base.y1
  const len = Math.hypot(dx, dy) || 1
  const nx = -dy / len
  const ny = dx / len
  const spread = (index - (total - 1) / 2) * gap
  return {
    x1: base.x1 + nx * spread,
    y1: base.y1 + ny * spread,
    x2: base.x2 + nx * spread,
    y2: base.y2 + ny * spread,
  }
}
