import { Prisma, copy_configurations, facebook_audiences, landing_pages, project_facebook_creative_templates, projects, projects_themes, teams_projects, themes_angles } from "@prisma/client"
import prisma from "./database"
import { v4 as uuidv4 } from 'uuid';
const getDemoTemplate = ({ projectId, teamId }: { projectId: string, teamId: string }): Prisma.PrismaPromise<any>[] => {
  const theme1Id = uuidv4()
  const theme2Id = uuidv4();
  const theme3Id = uuidv4()
  const startDate = new Date()
  const stopDate = new Date(startDate)
  stopDate.setDate(startDate.getDate() + 3)
  const project = { id: projectId, name: 'Demo Template', start_time: startDate, stop_time: stopDate, objective: 'Narrow down new flavors of chapstick', branding: 'unbranded', status: 'draft', platform: 'facebook_instagram', project_type: 'concept_test', product_description: 'This is a short description of my product' } as projects
  const teamProject = { project_id: projectId, team_id: teamId }
  const audience = { name: 'Gen Pop', project_id: projectId, min_age: 18, max_age: 65, genders: [1, 2], device_platforms: ["desktop", "mobile"], facebook_positions: ["feed"], geo_locations: { "regions": {}, "countries": ["US", "JP"] }, publisher_platforms: ["facebook", "instagram"], interests: [{ "id": "6003481451064", "name": "ChapStick" }] }
  const creativeTemplate = {
    project_id: projectId, template_name: 'LifestyleTemplate', data: {
      "mainCopy": "Balance your screentime.",
      "logoImage": null,
      "background": "#edfdf0"
    }
  }
  const themes = [
    { id: theme1Id, name: 'Value', project_id: projectId },
    { id: theme2Id, name: 'Ease of use', project_id: projectId },
    { id: theme3Id, name: 'Efficacy', project_id: projectId }
  ] as projects_themes[]
  const landingPage = {
    project_id: projectId,
    template_name: 'LandingPageSimplePoll',
    data: {
      "version": 1,
      "questions": [],
      "textColor": "",
      "submittedText": "",
      "headerImageUrl": "",
      "submitButtonText": "asdfasdf",
      "pageBackgroundColor": "",
      "submitButtonTextColor": "",
      "submitButtonBackgroundColor": ""
    }
  }
  const angles = [
    { theme_id: theme1Id, name: 'More bang for your buck' },
    { theme_id: theme1Id, name: 'Inexpensive alternative' },
    { theme_id: theme1Id, name: 'Save your money' },
    { theme_id: theme2Id, name: 'It\s easy' },
    { theme_id: theme2Id, name: 'Simple / Straightforward' },
    { theme_id: theme2Id, name: 'Familiar' },
    { theme_id: theme3Id, name: 'Good at what it does' },
    { theme_id: theme3Id, name: 'Made for a single purpose' },
    { theme_id: theme3Id, name: 'Quality' }
  ] as themes_angles[]
  const copyConfiguration = { project_id: projectId, brand_tone: 'Dry, fit, clean', character_count: 150, perspective: '3rd', tone: 'motivational', template_type: 'statement' } as copy_configurations
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