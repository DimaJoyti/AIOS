import { NextRequest, NextResponse } from 'next/server'

const AI_SERVICE_URL = process.env.AI_SERVICE_URL || 'http://localhost:8182'

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    
    // Validate request body
    if (!body.messages || !Array.isArray(body.messages)) {
      return NextResponse.json(
        { error: 'Messages array is required' },
        { status: 400 }
      )
    }

    // Forward request to AI service
    const response = await fetch(`${AI_SERVICE_URL}/api/v1/ai/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': request.headers.get('Authorization') || '',
      },
      body: JSON.stringify({
        messages: body.messages,
        model_id: body.model_id || 'gpt-3.5-turbo',
        system_prompt: body.system_prompt,
        config: {
          temperature: body.config?.temperature || 0.7,
          max_tokens: body.config?.max_tokens || 1000,
          top_p: body.config?.top_p || 1.0,
          frequency_penalty: body.config?.frequency_penalty || 0,
          presence_penalty: body.config?.presence_penalty || 0,
          stop_sequences: body.config?.stop_sequences || [],
        },
        user_id: body.user_id || 'anonymous',
        session_id: body.session_id,
        metadata: {
          source: 'web_frontend',
          timestamp: new Date().toISOString(),
          ...body.metadata
        }
      })
    })

    if (!response.ok) {
      const errorData = await response.text()
      console.error('AI service error:', errorData)
      return NextResponse.json(
        { error: 'AI service unavailable' },
        { status: 503 }
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
      created_at: data.created_at,
      metadata: data.metadata
    })

  } catch (error) {
    console.error('Chat API error:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const sessionId = searchParams.get('session_id')
    
    if (!sessionId) {
      return NextResponse.json(
        { error: 'Session ID is required' },
        { status: 400 }
      )
    }

    // Get chat history from session
    const response = await fetch(`${AI_SERVICE_URL}/api/v1/sessions/${sessionId}/history`, {
      headers: {
        'Authorization': request.headers.get('Authorization') || '',
      }
    })

    if (!response.ok) {
      return NextResponse.json(
        { error: 'Failed to fetch chat history' },
        { status: response.status }
      )
    }

    const data = await response.json()
    return NextResponse.json(data)

  } catch (error) {
    console.error('Chat history API error:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
