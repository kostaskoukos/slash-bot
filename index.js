require('dotenv').config();
const fs = require('fs');
const path = require('path');

const { Client, Events, GatewayIntentBits, Collection, ActionRowBuilder, ButtonBuilder, ButtonStyle, EmbedBuilder } = require('discord.js');
const { REST, Routes } = require('discord.js');
const { Player } = require('discord-player');

process.on('unhandledRejection', error => {
  console.error('Unhandled promise rejection:', error);
});

const client = new Client({ intents: [GatewayIntentBits.Guilds, GatewayIntentBits.GuildMessages, GatewayIntentBits.MessageContent, GatewayIntentBits.GuildVoiceStates,] });
const rest = new REST({ version: '10' }).setToken(process.env.token);

const commands = [];
client.commands = new Collection();

const filepath = path.join(__dirname, 'commands');
const files = fs.readdirSync(filepath).filter(file => file.endsWith('.js'));

for (const file of files) {
  const filePath = path.join(filepath, file);
  const cmd = require(filePath);

  if ('data' in cmd && 'run' in cmd) {
    commands.push(cmd.data.toJSON());
    client.commands.set(cmd.data.name, cmd);
  } else {
    console.log(`[WARNING] The command at ${filePath} is missing a required "data" or "run" property.`);
  }
}

rest.put(Routes.applicationCommands('1036229560376754247'), { body: commands })
  .then(data => console.log(`Successfully registered ${data.length} application commands.`));

///////////////////////////////////////////   VOICE    ///////////////////////////
const player = new Player(client, {
  initialVolume: 100, ytdlOptions: { quality: 'highestaudio' }
});
player.on('error', (q, e) => { console.log(e) });
player.on('connectionError', (q, e) => { console.log(e) });
client.player = player;
client.playlist = [];
client.now = {};//own implementation
///////////////////////////////////////////   VOICE    ///////////////////////////

const makelistembed = () => {
  let list = [];
  client.playlist.forEach((song, i) => {
    const playing = song == client.now;
    list.push(`${playing ? `:arrow_forward:__[${song.title}](${song.url})__` : `[${song.title}](${song.url})`} [${i + 1}] \n`);
  });
  return new EmbedBuilder()
    .setColor(0x404EED)
    .setDescription(`**${list.join('')}**`);
}

const row = new ActionRowBuilder()
  .addComponents(
    new ButtonBuilder()
      .setCustomId('left')
      .setLabel('◀️')
      .setStyle(ButtonStyle.Secondary),
    new ButtonBuilder()
      .setCustomId('right')
      .setLabel('▶️')
      .setStyle(ButtonStyle.Secondary));


client.on(Events.InteractionCreate, async msg => {
  await msg.deferReply();
  if (msg.isButton()) {
    const queue = msg.client.player.getQueue(msg.guild);
    switch (msg.customId) {
      case 'left':
        if (msg.client.playlist.indexOf(msg.client.now) == 0) return await msg.reply('Δεν υπάρχει προηγούμενο τραγούδι');//if first song
        const lsong = msg.client.playlist[msg.client.playlist.indexOf(msg.client.now) - 1];
        queue.play(lsong, { immediate: true });
        msg.client.now = lsong;
        await msg.editReply(`Τώρα παίζει **[${lsong.title}](${lsong.url})**`);
        await msg.followUp({ embeds: [makelistembed()], components: [row] });
        return;
      case 'right':
        if (msg.client.playlist.length - 1 == msg.client.playlist.indexOf(msg.client.now)) return await msg.reply('Δεν υπάρχει επόμενο τραγούδι');//if last song
        const rsong = msg.client.playlist[msg.client.playlist.indexOf(msg.client.now) + 1];
        queue.play(rsong, { immediate: true });
        msg.client.now = rsong;
        await msg.editReply(`Τώρα παίζει **[${rsong.title}](${rsong.url})**`);
        await msg.followUp({ embeds: [makelistembed()], components: [row] });
        return;
    }
  }

  if (!msg.isChatInputCommand()) return;

  const cmd = client.commands.get(msg.commandName);
  console.log(cmd);

  if (!cmd) {
    console.error(`No command matching ${msg.commandName} was found.`);
    return;
  }

  try {
    return await cmd.run(msg);
  } catch (error) {
    console.log(error);
  }
});

client.login(process.env.token);
///////////////////////////////////////    KEEP ALIVE    //////////////////////////////////

// const http = require('http');

// http.createServer((req, res) => {
//   res.write("alive");
//   res.end();
// }).listen(8080);

// client.on('ready', () => {
//   console.log('bot alive')
//   setInterval(() => client.user.setActivity(''), 5000);
// });
