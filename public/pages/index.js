
function credsWidget(folders, entries, search) {
	var credsTable = createTable();

	var sw = searchWidget(search);

	var titleTh = credsTable.th().text('Title');
	credsTable.th().text('Username');

	titleTh.append([ document.createElement('br'), sw ]);

	for (var i = 0; i < folders.length; ++i) {
		var sub = folders[i];

		var tr = credsTable.tr();
		var folderTitleTd = credsTable.td(tr);

		var folderIcon = $('<span class="glyphicon glyphicon-folder-open"></span>');

		$('<a></a>')
			.attr('href', linkTo([ 'index', sub.Id ]))
			.append([ folderIcon, ' &nbsp;', sub.Name ])
			.appendTo(folderTitleTd);

		credsTable.td(tr);
	}

	for (var i = 0; i < entries.length; ++i) {
		var tr = credsTable.tr();

		var titleTd = credsTable.td(tr);
		$('<a></a>')
			.attr('href', linkTo([ 'credview', entries[i].Id ]))
			.text(entries[i].Title)
			.appendTo(titleTd);

		credsTable.td(tr).text(entries[i].Username);
	}

	return credsTable.table;
}

routes.index = function(args) {
	var folderId = args[1] || 'root';

	rest_byFolder(folderId).then(function (resp) {
		var bcItems = resp.ParentFolders.reverse().map(function (item){
			return {
				href: linkTo([ 'index', item.Id ]),
				label: item.Name
			};
		});

		bcItems.push({
			href: '', // don't need one - last item is not linked to
			label: resp.Folder.Name
		});

		breadcrumbWidget(bcItems).appendTo(cc());

		credsWidget(resp.SubFolders, resp.Accounts, null).appendTo(cc());

		var secretCreateBtn = $('<button class="btn btn-default"></button>')
			.text('+ Account')
			.appendTo(cc());

		attachCommand(secretCreateBtn, {
			cmd: 'SecretCreateRequest',
			prefill: {
				FolderId: folderId
			} });

		var folderCreateBtn = $('<button class="btn btn-default"></button>')
			.text('+ Folder')
			.appendTo(cc());

		attachCommand(folderCreateBtn, {
			cmd: 'FolderCreateRequest',
			prefill: {
				ParentId: folderId
			} });

		var folderRenameBtn = $('<button class="btn btn-default"></button>')
			.text('Rename folder')
			.appendTo(cc());

		attachCommand(folderRenameBtn, {
			cmd: 'RenameFolderRequest',
			prefill: {
				Id: resp.Folder.Id,
				Name: resp.Folder.Name
			} });

		var folderRenameBtn = $('<button class="btn btn-default"></button>')
			.text('Move folder')
			.appendTo(cc());

		attachCommand(folderRenameBtn, {
			cmd: 'MoveFolderRequest',
			prefill: {
				Id: resp.Folder.Id,
				ParentId: resp.Folder.ParentId
			} });
	}, restDefaultErrorHandler);
}
