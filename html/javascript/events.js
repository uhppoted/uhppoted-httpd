import { DB } from './db.js'

export function refreshed () {
  const events = [...DB.events.values()] // .sort((p, q) => p.index - q.index)

  console.log('EVENTS', events)

  // realize(groups)

  // groups.forEach(o => {
  //   const row = updateFromDB(o.OID, o)
  //   if (row) {
  //     if (o.status === 'new') {
  //       row.classList.add('new')
  //     } else {
  //       row.classList.remove('new')
  //     }
  //   }
  // })

  // DB.refreshed('groups')
}
