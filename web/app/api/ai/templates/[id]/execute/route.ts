import { NextRequest, NextResponse } from 'next/server'

const AI_SERVICE_URL = process.env.AI_SERVICE_URL || 'http://localhost:8182'

export async function POST(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    const templateId = params.id
    const body = await request.json()
    
    // Validate request body
    if (!body.variables || typeof body.variables !== 'object') {
      return NextResponse.json(
        { error: 'Variables object is required' },
        { status: 400 }
      )
    }

    // Forward request to AI service
    const response = await fetch(`${AI_SERVICE_URL}/api/v1/ai/templates/${templateId}/execute`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': request.headers.get('Authorization') || '',
      },
      body: JSON.stringify({
        variables: body.variables,
        user_id: body.user_id || 'anonymous',
        session_id: body.session_id,
        metadata: {
          source: 'web_frontend',
          timestamp: new Date().toISOString(),
          template_id: templateId,
          ...body.metadata
        }
      })
    })

    if (!response.ok) {
      const errorData = await response.text()
      console.error('AI service error:', errorData)
      
      if (response.status === 404) {
        return NextResponse.json(
          { error: 'Template not found' },
          { status: 404 }
        )
      }
      
      return NextResponse.json(
        { error: 'Failed to execute template' },
        { status: response.status }
      )
    }

    const data = await response.json()
    
    return NextResponse.json({
      id: data.id,
      text: data.text,
      finish_reason: data.finish_reason,
      usage: data.usage,
      cost: data.cost,
      latency: data.latency,
      model_id: data.model_id,
      provider: data.provider,
      template_id: templateId,
      variables: body.variables,
      created_at: data.created_at,
      metadata: data.metadata
    })

  } catch (error) {
    console.error('Template execution API error:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
