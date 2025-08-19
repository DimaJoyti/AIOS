import { NextRequest, NextResponse } from 'next/server'

const KNOWLEDGE_SERVICE_URL = process.env.KNOWLEDGE_SERVICE_URL || 'http://localhost:8081'

export async function POST(request: NextRequest) {
  try {
    const formData = await request.formData()
    const file = formData.get('file') as File
    
    if (!file) {
      return NextResponse.json(
        { error: 'No file provided' },
        { status: 400 }
      )
    }

    // Validate file type
    const allowedTypes = [
      'application/pdf',
      'application/msword',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      'text/plain',
      'text/markdown',
      'application/rtf'
    ]

    if (!allowedTypes.includes(file.type)) {
      return NextResponse.json(
        { error: 'Unsupported file type' },
        { status: 400 }
      )
    }

    // Validate file size (max 10MB)
    const maxSize = 10 * 1024 * 1024 // 10MB
    if (file.size > maxSize) {
      return NextResponse.json(
        { error: 'File too large. Maximum size is 10MB' },
        { status: 400 }
      )
    }

    // Create form data for knowledge service
    const knowledgeFormData = new FormData()
    knowledgeFormData.append('file', file)
    knowledgeFormData.append('metadata', JSON.stringify({
      source: 'web_frontend',
      uploaded_at: new Date().toISOString(),
      user_id: request.headers.get('X-User-ID') || 'anonymous'
    }))

    // Forward to knowledge service
    const response = await fetch(`${KNOWLEDGE_SERVICE_URL}/api/v1/documents/upload`, {
      method: 'POST',
      headers: {
        'Authorization': request.headers.get('Authorization') || '',
      },
      body: knowledgeFormData
    })

    if (!response.ok) {
      const errorData = await response.text()
      console.error('Knowledge service error:', errorData)
      return NextResponse.json(
        { error: 'Failed to upload document' },
        { status: response.status }
      )
    }

    const data = await response.json()
    
    // Return processed document information
    return NextResponse.json({
      id: data.id,
      name: file.name,
      type: file.type,
      size: file.size,
      status: data.status || 'processing',
      summary: data.summary,
      extractedText: data.extracted_text,
      metadata: {
        pages: data.metadata?.pages,
        wordCount: data.metadata?.word_count,
        language: data.metadata?.language,
        ...data.metadata
      },
      suggestedTags: data.suggested_tags || [],
      uploadedAt: new Date().toISOString()
    })

  } catch (error) {
    console.error('Document upload API error:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const page = parseInt(searchParams.get('page') || '1')
    const limit = parseInt(searchParams.get('limit') || '20')
    const search = searchParams.get('search')
    const tags = searchParams.get('tags')
    
    let url = `${KNOWLEDGE_SERVICE_URL}/api/v1/documents`
    const params = new URLSearchParams()
    
    params.append('page', page.toString())
    params.append('limit', limit.toString())
    if (search) params.append('search', search)
    if (tags) params.append('tags', tags)
    
    url += `?${params.toString()}`

    const response = await fetch(url, {
      headers: {
        'Authorization': request.headers.get('Authorization') || '',
      }
    })

    if (!response.ok) {
      return NextResponse.json(
        { error: 'Failed to fetch documents' },
        { status: response.status }
      )
    }

    const data = await response.json()
    return NextResponse.json(data)

  } catch (error) {
    console.error('Documents list API error:', error)
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}
