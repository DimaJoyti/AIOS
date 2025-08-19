import { NextRequest, NextResponse } from 'next/server'

const AI_SERVICE_URL = process.env.AI_SERVICE_URL || 'http://localhost:8182'

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const category = searchParams.get('category')
    const tags = searchParams.get('tags')
    
    let url = `${AI_SERVICE_URL}/api/v1/ai/templates`
    const params = new URLSearchParams()
    
    if (category) params.append('category', category)
    if (tags) params.append('tags', tags)
    
    if (params.toString()) {
      url += `?${params.toString()}`
    }

    const response = await fetch(url, {
      headers: {
        'Authorization': request.headers.get('Authorization') || '',
      }
    })

    if (!response.ok) {
      return NextResponse.json(
        { error: 'Failed to fetch templates' },
        { status: response.status }
      )
    }

    const data = await response.json()
    return NextResponse.json(data)

  } catch (error) {
    console.error('Templates API error:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    
    // Validate required fields
    if (!body.name || !body.template) {
      return NextResponse.json(
        { error: 'Name and template content are required' },
        { status: 400 }
      )
    }

    // Forward request to AI service
    const response = await fetch(`${AI_SERVICE_URL}/api/v1/ai/templates`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': request.headers.get('Authorization') || '',
      },
      body: JSON.stringify({
        id: body.id || `template_${Date.now()}`,
        name: body.name,
        description: body.description || '',
        category: body.category || 'custom',
        template: body.template,
        variables: body.variables || [],
        examples: body.examples || [],
        config: {
          model_id: body.config?.model_id || 'gpt-3.5-turbo',
          temperature: body.config?.temperature || 0.7,
          max_tokens: body.config?.max_tokens || 1000,
          top_p: body.config?.top_p || 1.0,
          frequency_penalty: body.config?.frequency_penalty || 0,
          presence_penalty: body.config?.presence_penalty || 0,
          stop_sequences: body.config?.stop_sequences || [],
          system_prompt: body.config?.system_prompt || '',
          parameters: body.config?.parameters || {}
        },
        tags: body.tags || [],
        version: body.version || '1.0',
        created_by: body.created_by || 'user',
        metadata: {
          source: 'web_frontend',
          created_at: new Date().toISOString(),
          ...body.metadata
        }
      })
    })

    if (!response.ok) {
      const errorData = await response.text()
      console.error('AI service error:', errorData)
      return NextResponse.json(
        { error: 'Failed to create template' },
        { status: response.status }
      )
    }

    const data = await response.json()
    return NextResponse.json(data)

  } catch (error) {
    console.error('Create template API error:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
