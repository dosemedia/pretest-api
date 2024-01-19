import { Prisma, copy_configurations, facebook_audiences, landing_pages, project_facebook_creative_templates, projects, projects_themes, teams_projects, themes_angles } from "@prisma/client"
import prisma from "./database"
import { v4 as uuidv4 } from 'uuid';
const getDemoTemplate = ({ projectId, teamId }: { projectId: string, teamId: string }): Prisma.PrismaPromise<any>[] => {
  const theme1Id = uuidv4()
  const theme2Id = uuidv4();
  const theme3Id = uuidv4()
  const project = { id: projectId, name: 'Hydrogen Infused Water Bottle Product Concept Exploration', duration: 3, objective: 'I want to understand how to bring this product idea to life by determining the most resonating concepts and identify my target audience for future testing iterations', branding: 'unbranded', status: 'draft', platform: 'facebook_instagram', project_type: 'concept_test', product_description: 'This water bottle with infuse hydrogen into your water for packed hydration.' } as projects
  const teamProject = { project_id: projectId, team_id: teamId }
  const audience = { name: 'Gen Pop', project_id: projectId, min_age: 18, max_age: 65, genders: [1, 2], device_platforms: ["desktop", "mobile"], facebook_positions: ["feed"], geo_locations: { "regions": {}, "countries": ["US"] }, publisher_platforms: ["facebook", "instagram"] }
  const creativeTemplate = {
    project_id: projectId, template_name: 'ProductTemplate', data: {
      "mainCopy": "Find the perfect on-the-go snack at your local bakery.",
      "productImage": null,
      "background": "#dce9be"
    }
  }

  const themes = [
    { id: theme1Id, name: 'Sustainability', project_id: projectId },
    { id: theme2Id, name: 'Authenticity', project_id: projectId },
    { id: theme3Id, name: 'Efficacy', project_id: projectId }

  ] as projects_themes[]
  const angles = [
    { theme_id: theme1Id, name: 'Eco-friendly alternative' },
    { theme_id: theme1Id, name: 'Saves precious resources' },
    { theme_id: theme1Id, name: 'Less waste' },
    { theme_id: theme2Id, name: 'Knows who you are' },
    { theme_id: theme2Id, name: 'Knows what your problems are' },
    { theme_id: theme2Id, name: 'Knows how to solve your problems' },
    { theme_id: theme3Id, name: 'Good at what it does' },
    { theme_id: theme3Id, name: 'Made for a single purpose' },
    { theme_id: theme3Id, name: 'Quality' }
  ] as themes_angles[]
  const landingPage = {
    project_id: projectId,
    template_name: 'LandingPageSimplePoll',
    data: {
      "version": 1,
      "questions": [
        {
          "title": "What is the price range you would consider purchasing a product like this?",
          "options": [
            "$10-20",
            "$20-30",
            "$30-40",
            "$50+"
          ],
          "multipleChoice": false
        },
        {
          "title": "What additonal features would you want for a product like this?",
          "options": [
            "Temperature Regulation",
            "Durability",
            "Cool Designs & Styles",
            "Ease of Cleaning",
            "Material Safety"
          ],
          "multipleChoice": true
        }
      ],
      "textColor": "#ffffff",
      "submittedText": "Thank you for submitting!",
      "headerImageUrl": "http://localhost:3000/files/project-assets/cc91c532-7cfd-4e80-9193-e1959831e5b8/landing_pages/d9a89f14-ce66-447e-8ca8-5b8271fbd17a.png",
      "submitButtonText": "SUBMIT",
      "pageBackgroundColor": "#6638e5",
      "submitButtonTextColor": "",
      "submitButtonBackgroundColor": ""
    }
  }

  const copyConfiguration = { project_id: projectId, brand_tone: 'science-backed, technology-focused, wellness-oriented', character_count: 50, perspective: '3rd', tone: 'motivational', template_type: 'statement' } as copy_configurations
  return [
    prisma.projects.create({ data: project }),
    prisma.teams_projects.create({ data: teamProject }),
    prisma.facebook_audiences.create({ data: audience }),
    prisma.projects_themes.createMany({ data: themes }),
    prisma.themes_angles.createMany({ data: angles }),
    prisma.landing_pages.create({ data: landingPage }),
    prisma.project_facebook_creative_templates.create({ data: creativeTemplate }),
    prisma.copy_configurations.create({ data: copyConfiguration })
  ]
}



export class ProjectTemplate {
  projectId: string
  teamId: string
  constructor({ projectId, teamId }: { projectId: string, teamId: string }) {
    this.projectId = projectId
    this.teamId = teamId;
  }

  useTemplate({ templateName }: { templateName: string }): Prisma.PrismaPromise<any>[] {
    switch (templateName) {
      case 'demo':
        return getDemoTemplate({ projectId: this.projectId, teamId: this.teamId })
    }
    return []
  }
}